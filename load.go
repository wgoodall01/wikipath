package main

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

var StoppedErr error = errors.New("Visitor stopped reading.")

func LoadWikiToIndex(source io.Reader, ind *Index) {
	// Does this in parallel. It's roughly 30% faster.

	loadChan := make(chan *Article, 100)

	go func() {
		LoadWiki(source, func(a *Article) bool {
			loadChan <- a
			return true
		})
		close(loadChan)
	}()

	for a := range loadChan {
		ind.AddArticle(a)
	}

}

const chanSize int = 1024       // Buffers inbetween all channels
const readerBufSize int = 50000 // File buffers in front of OS

type loadContext struct {
	wg       sync.WaitGroup
	Canceled chan struct{}

	Err    error
	errMut sync.Mutex

	latest int
}

func newLoadContext() *loadContext {
	lc := &loadContext{Canceled: make(chan struct{})}
	return lc
}

// Add(n) adds n workers to the loadContext.
func (lc *loadContext) Add(n int) {
	lc.wg.Add(n)
	lc.latest = lc.latest + n
}

// Start() adds a worker and returns an ID for it.
func (lc *loadContext) Start() int {
	lc.wg.Add(1)

	lc.latest++
	return lc.latest
}

// Done() marks a worker as done.
func (lc *loadContext) Done() {
	lc.wg.Done()
}

func (lc *loadContext) Cancel(err error) {
	lc.errMut.Lock()
	select {
	case <-lc.Canceled:
		// Do nothing, already canceled.
	default:
		// Not yet canceled.
		lc.Err = err
		close(lc.Canceled) // cancel the context.
	}
	lc.errMut.Unlock()
}

func (lc *loadContext) Wait() error {
	success := make(chan struct{})

	go func() {
		lc.wg.Wait()
		close(success)
	}()

	select {
	case <-success:
		return nil
	case <-lc.Canceled:
		return lc.Err
	}
}

func LoadWikiCompressed(index io.Reader, source io.ReaderAt, visitor func(*Article) bool) error {
	lc := newLoadContext()

	chunks := loadIndexChunks(lc, index)
	articles := make(chan *Article, chanSize)

	nWorkers := runtime.GOMAXPROCS(4)

	for i := 0; i < nWorkers; i++ {
		decompressChunks(lc, source, chunks, articles)
	}

	go func() {
		for a := range articles {
			shouldCont := visitor(a)

			if !shouldCont {
				lc.Cancel(StoppedErr)
			}
		}
	}()

	err := lc.Wait()
	close(articles)
	return err
}

func decompressChunks(lc *loadContext, archiveFile io.ReaderAt, chunks <-chan [2]int64, articles chan<- *Article) {
	lc.Start()
	go func() {
		for chunk := range chunks {
			chunkRaw := io.NewSectionReader(archiveFile, chunk[0], chunk[1]-chunk[0])
			chunkReader := bufio.NewReaderSize(chunkRaw, readerBufSize)
			chunkDecompressor := bzip2.NewReader(chunkReader)
			loadErr := LoadWiki(chunkDecompressor, func(a *Article) bool {
				select {
				case <-lc.Canceled:
					// Load has been canceled
					return false
				default:
					articles <- a
					// Still loading
					return true
				}
			})

			if loadErr != nil {
				lc.Cancel(loadErr)
				return
			}
		}
		lc.Done()
	}()
}

func loadIndexChunks(lc *loadContext, indexRaw io.Reader) <-chan [2]int64 {
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
				lc.Cancel(parseErr)
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
			lc.Cancel(err)
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
