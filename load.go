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

func LoadWikiToIndex(source io.Reader, ind *Index) {
	// Does this in parallel. It's roughly 30% faster.

	loadChan := make(chan *Article, 100)

	go func() {
		LoadWiki(source, func(a *Article) {
			loadChan <- a
		})
		close(loadChan)
	}()

	for a := range loadChan {
		ind.AddArticle(a)
	}

}

const chanSize int = 1024       // Buffers inbetween all channels
const readerBufSize int = 50000 // File buffers in front of OS

func LoadWikiCompressed(index io.Reader, source io.ReaderAt, visitor func(*Article)) {

	chunks := loadIndexChunks(index)
	articles := make(chan *Article, chanSize)

	nWorkers := runtime.GOMAXPROCS(4)
	var wg sync.WaitGroup
	wg.Add(nWorkers)

	for i := 0; i < nWorkers; i++ {
		decompressChunks(&wg, source, chunks, articles)
	}

	go func() {
		wg.Wait()
		close(articles)
	}()

	for a := range articles {
		visitor(a)
	}
}

func decompressChunks(wg *sync.WaitGroup, archiveFile io.ReaderAt, chunks <-chan [2]int64, articles chan<- *Article) {
	go func() {
		for chunk := range chunks {
			chunkRaw := io.NewSectionReader(archiveFile, chunk[0], chunk[1]-chunk[0])
			chunkReader := bufio.NewReaderSize(chunkRaw, readerBufSize)
			chunkDecompressor := bzip2.NewReader(chunkReader)
			LoadWiki(chunkDecompressor, func(a *Article) {
				articles <- a
			})
		}
		wg.Done()
	}()
}

func loadIndexChunks(indexRaw io.Reader) <-chan [2]int64 {
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
			chunk, _, _, _ := ParseIndexLine(line)
			//TODO: handle error from ParseIndexLine
			if chunk > offset {
				chunks <- [2]int64{offset, chunk}
				offset = chunk
			}
		}
		close(chunks)
	}()

	return chunks
}

func LoadWiki(source io.Reader, visitor func(*Article)) {
	// Open an XML decoder over the file.
	decoder := xml.NewDecoder(source)

	for {
		// Get the next token.
		tok, _ := decoder.Token()
		if tok == nil {
			break
		}
		//TODO: handle errors from token

		switch se := tok.(type) {
		case xml.StartElement:
			// Element is a starting element
			if se.Name.Local == "page" {
				var a Article
				decoder.DecodeElement(&a, &se)

				visitor(&a)
			}
		}
	}
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
