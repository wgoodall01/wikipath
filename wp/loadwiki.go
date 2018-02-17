package wikipath

import (
	"bufio"
	"compress/bzip2"
	"encoding/xml"
	"errors"
	"io"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var StoppedErr error = errors.New("Visitor stopped reading.")

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Article struct {
	Id                int      `xml:"id"`
	Namespace         int      `xml:"ns"`
	Title             string   `xml:"title"`
	Redirect          Redirect `xml:"redirect"`
	Text              string   `xml:"revision>text"`
	RevisionId        int      `xml:"revision>id"`
	RevisionTimestamp string   `xml:"revision>timestamp"`
	RevisionFormat    string   `xml:"revision>format"`
	RevisionAuthor    string   `xml:"revision>contributor>username"`
	RevisionAuthorId  string   `xml:"revision>contributor>id"`
}

const chanSize int = 1024       // Buffers inbetween all channels
const readerBufSize int = 50000 // File buffers in front of OS

func LoadWikiCompressed(index io.Reader, source io.ReaderAt, visitor func(*Article) bool) error {
	ec := NewErrorContext()

	chunks := loadIndexChunks(ec, index)
	articles := make(chan *Article, chanSize)

	nWorkers := runtime.GOMAXPROCS(-1)

	for i := 0; i < nWorkers; i++ {
		decompressChunks(ec, source, chunks, articles)
	}

	var done sync.WaitGroup
	go func() {
		done.Add(1)
		for a := range articles {
			shouldCont := visitor(a)

			if !shouldCont {
				ec.Cancel(StoppedErr)
			}
		}
		done.Done()
	}()

	ec.Wait()
	close(articles)
	done.Wait()
	return ec.Err
}

func decompressChunks(ec *ErrorContext, archiveFile io.ReaderAt, chunks <-chan [2]int64, articles chan<- *Article) {
	ec.Start()
	go func() {
		for chunk := range chunks {
			chunkRaw := io.NewSectionReader(archiveFile, chunk[0], chunk[1]-chunk[0])
			chunkReader := bufio.NewReaderSize(chunkRaw, readerBufSize)
			chunkDecompressor := bzip2.NewReader(chunkReader)
			loadErr := LoadWiki(chunkDecompressor, func(a *Article) bool {
				select {
				case <-ec.Canceled:
					// Load has been canceled
					return false
				default:
					articles <- a
					// Still loading
					return true
				}
			})

			if loadErr != nil {
				ec.Cancel(loadErr)
				return
			}
		}
		ec.Done()
	}()
}

func loadIndexChunks(ec *ErrorContext, indexRaw io.Reader) <-chan [2]int64 {
	chunks := make(chan [2]int64, chanSize)

	// Open a decompressing reader on indexPath
	indexBuf := bufio.NewReaderSize(indexRaw, readerBufSize)
	indexReader := bzip2.NewReader(indexBuf)
	indexScanner := bufio.NewScanner(indexReader)

	go func() {
		// Load the first line, get the first offset
		var offset int64 = 0

		for indexScanner.Scan() {
			line := indexScanner.Text()
			chunk, _, _, parseErr := ParseIndexLine(line)
			if parseErr != nil {
				ec.Cancel(parseErr)
				return
			}
			if chunk > offset {
				chunks <- [2]int64{offset, chunk}
				offset = chunk
			}
		}

		close(chunks)

		err := indexScanner.Err()
		if err != nil {
			ec.Cancel(err)
			return
		}
	}()

	return chunks
}

func LoadWiki(source io.Reader, visitor func(*Article) bool) error {
	// Open an XML decoder over the file.
	decoder := xml.NewDecoder(source)

	for {
		// Get the next token.
		tok, tokErr := decoder.Token()
		if tok == nil {
			return nil
		}

		if tokErr != nil {
			return tokErr
		}

		switch se := tok.(type) {
		case xml.StartElement:
			// Element is a starting element
			if se.Name.Local == "page" {
				var a Article
				decoder.DecodeElement(&a, &se)

				shouldCont := visitor(&a)

				if !shouldCont {
					return StoppedErr
				}
			}
		}
	}

	return nil
}

// https://regex101.com/r/Q2bNwC/3
var LinkRegex = regexp.MustCompile(`(?U)\[\[([^]:]+)([#/|].+)?\]\]`)

// Returns a list of strings, representing the titles of articles.
func ParseLinks(text string) []string {
	matches := LinkRegex.FindAllStringSubmatchIndex(text, -1)
	linkNames := make([]string, len(matches))
	for i, inds := range matches {
		start := inds[2]
		end := inds[3]
		linkNames[i] = text[start:end]
	}
	return linkNames
}

func ParseIndexLine(line string) (int64, uint, string, error) {
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, ":", 3)

	offset, err0 := strconv.ParseInt(parts[0], 10, 64)
	if err0 != nil {
		return 0, 0, "", errors.New("Failed to parse offset")
	}

	id64, err1 := strconv.ParseUint(parts[1], 10, 0)
	if err1 != nil {
		return 0, 0, "", errors.New("Failed to parse ID")
	}
	id := uint(id64)

	name := parts[2]

	return offset, id, name, nil
}
