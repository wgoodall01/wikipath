package main

import (
	"bufio"
	"compress/bzip2"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

var archivePath = flag.String("archivePath", "./wikis/simple.xml", "Path to the wiki dump, as an .xml file.")
var bzipPath = flag.String("bzipPath", "./wikis/simple.xml.bz2", "Path to the wiki dump, as an .xml.bz2 file.")
var indexPath = flag.String("indexPath", "./wikis/simple-index.txt", "Path to the index, as a .txt")
var bzipIndexPath = flag.String("bzipIndexPath", "./wikis/simple-index.txt.bz2", "Path to the index, as a .txt.bz2")

func TestIndexedGzip(t *testing.T) {
	archivePath := "./wikis/simple.xml.bz2"
	//indexPath := "./wikis/simple-index.txt.bz2"

	t.Log("Testing indexed bzip loading...")

	bzipFile, _ := os.Open(archivePath)
	archiveReader := bzip2.NewReader(bzipFile)
	LoadWiki(archiveReader, func(a *Article) error {
		t.Logf("Article: %s", a.Title)
		return errors.New("")
	})
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		msg := fmt.Sprintf("%v != %v", a, b)
		t.Fatal(msg)
	}
}

const testXml = `
<mediawiki>
  <page>
    <title>Abrahamic religion</title>
   	<redirect title="Testing redirect title"/>
	<ns>0</ns>
    <id>43</id>
    <revision>
      <id>5647518</id>
      <parentid>5647517</parentid>
      <timestamp>2017-03-29T18:45:41Z</timestamp>
      <contributor>
        <username>Tegel</username>
        <id>67822</id>
      </contributor>
      <minor />
      <comment>[[Help:Revert a page|Reverted]] edits by [[Special:Contributions/198.147.198.221|198.147.198.221]] ([[User talk:198.147.198.221|talk]]) to last version by 61.69.102.70</comment>
      <model>wikitext</model>
      <format>text/x-wiki</format>
	  <text>This is some [[example]] text.</text>
      <sha1>6j946t5bta8mev2lxm1canivpsibwbw</sha1>
    </revision>
  </page>
</mediawiki>
`

func TestGetAnArticle(t *testing.T) {
	xmlReader := strings.NewReader(testXml)

	cb := func(a *Article) error {
		assertEqual(t, a.Title, "Abrahamic religion")
		assertEqual(t, a.Redirect.Title, "Testing redirect title")
		assertEqual(t, a.Text, "This is some [[example]] text.")
		assertEqual(t, a.Namespace, 0)
		return nil
	}

	LoadWiki(xmlReader, cb)

	t.Log("Done.")
}

const testWikitext = `
[[Sandbox]]
[[Fox Broadcasting Company|Fox]]
[[Queen (band)|Queen]]
[[Queen (chess)|Queen]]
[[Target page#Target section|display text]]
[[Wikipedia:Tutorial/Wikipedia_links#Categories|Categories]]
''[[War and Peace]]''
[[Image:Addition.gif|thumb|220px|Addition ]]
[[Cilk]] – a concurrent [[C (programming language)|C]]
`

func TestParseLinks(t *testing.T) {
	links := ParseLinks(testWikitext)
	assertEqual(t, links[0], "Sandbox")
	assertEqual(t, links[1], "Fox Broadcasting Company")
	assertEqual(t, links[2], "Queen (band)")
	assertEqual(t, links[3], "Queen (chess)")
	assertEqual(t, links[4], "Target page")
	assertEqual(t, links[5], "War and Peace")
	assertEqual(t, links[6], "Cilk")
	assertEqual(t, links[7], "C (programming language)")
}

func checkError(b *testing.B, err error) {
	if err != nil {
		b.Fatal(err.Error())
	}
}

func BenchmarkLoadXML(b *testing.B) {

	b.Run("LoadSync", func(b *testing.B) {
		archiveFile, fileErr := os.Open(*archivePath)
		checkError(b, fileErr)

		ind := NewIndex()
		LoadWiki(archiveFile, func(a *Article) error {
			ind.AddArticle(a)
			return nil
		})
	})

	b.Run("LoadBzippedSync", func(b *testing.B) {
		bzipRaw, fileErr := os.Open(*bzipPath)
		checkError(b, fileErr)
		bzipFile := bufio.NewReader(bzipRaw)
		archiveStream := bzip2.NewReader(bzipFile)

		ind := NewIndex()
		LoadWiki(archiveStream, func(a *Article) error {
			ind.AddArticle(a)
			return nil
		})
	})

	b.Run("LoadAsync", func(b *testing.B) {
		archiveFile, fileErr := os.Open(*archivePath)
		checkError(b, fileErr)
		ind := NewIndex()

		loadChan := make(chan *Article, 100)

		go func() {
			LoadWiki(archiveFile, func(a *Article) error {
				loadChan <- a
				return nil
			})
			close(loadChan)
		}()

		for a := range loadChan {
			ind.AddArticle(a)
		}

	})

	b.Run("LoadBzippedAsync", func(b *testing.B) {
		chanSize := 64         // Buffers inbetween all channels
		readerBufSize := 10000 // File buffers in front of OS

		parseIndexLine := func(line string) (int64, uint, string) {
			line = strings.TrimSpace(line)
			parts := strings.SplitN(line, ":", 3)

			offset, err0 := strconv.ParseInt(parts[0], 10, 64)
			checkError(b, err0)

			id64, err1 := strconv.ParseUint(parts[1], 10, 0)
			checkError(b, err1)
			id := uint(id64)

			name := parts[2]

			return offset, id, name
		}

		loadIndexChunks := func(indexPath string, stop *bool) <-chan [2]int64 {
			// Create the chunk channel
			chunks := make(chan [2]int64, chanSize)

			go func() {
				// Open a decompressing reader on indexPath
				indexRaw, indexErr := os.Open(indexPath)
				checkError(b, indexErr)
				indexBuf := bufio.NewReaderSize(indexRaw, readerBufSize)
				indexReader := bzip2.NewReader(indexBuf)
				indexScanner := bufio.NewScanner(indexReader)

				// Load the first line, get the first offset
				var offset int64 = 0

				for indexScanner.Scan() {
					line := indexScanner.Text()
					chunk, _, _ := parseIndexLine(line)
					if chunk > offset {
						if *stop {
							return
						} else {
							chunks <- [2]int64{offset, chunk}
							offset = chunk
						}
					}
				}
				close(chunks)
			}()

			return chunks
		}

		decompressSections := func(archivePath string, chunks <-chan [2]int64) <-chan *Article {
			articles := make(chan *Article, chanSize)

			go func() {
				for chunk := range chunks {
					archiveFile, archiveErr := os.Open(archivePath)
					checkError(b, archiveErr)
					chunkRaw := io.NewSectionReader(archiveFile, chunk[0], chunk[1]-chunk[0])
					chunkReader := bufio.NewReaderSize(chunkRaw, readerBufSize)
					chunkDecompressor := bzip2.NewReader(chunkReader)
					LoadWiki(chunkDecompressor, func(a *Article) error {
						articles <- a
						return nil
					})
				}
				close(articles)
			}()

			return articles
		}

		mergeArticles := func(chans ...<-chan *Article) <-chan *Article {
			var wg sync.WaitGroup
			out := make(chan *Article, chanSize)

			wg.Add(len(chans))
			for _, ch := range chans {
				go func() {
					for a := range ch {
						out <- a
					}
					wg.Done()
				}()
			}

			go func() {
				wg.Wait()
				close(out)
			}()

			return out

		}

		stop := false
		ind := NewIndex()
		chunks := loadIndexChunks(*bzipIndexPath, &stop)

		nWorkers := runtime.GOMAXPROCS(0)
		articleChans := make([]<-chan *Article, nWorkers)
		for i := 0; i < nWorkers; i++ {
			articleChans[i] = decompressSections(*bzipPath, chunks)
		}

		for a := range mergeArticles(articleChans...) {
			ind.AddArticle(a)
		}

		stop = true

	})

}
