package wikipath

import (
	"bufio"
	"compress/bzip2"
	"fmt"
	"os"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		msg := fmt.Sprintf("%v != %v", a, b)
		t.Fatal(msg)
	}
}

const testXML = `
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
	xmlReader := strings.NewReader(testXML)

	cb := func(a *Article) bool {
		assertEqual(t, a.Title, "Abrahamic religion")
		assertEqual(t, a.Redirect.Title, "Testing redirect title")
		assertEqual(t, a.Text, "This is some [[example]] text.")
		assertEqual(t, a.Namespace, 0)
		return true
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
		archiveFile, fileErr := os.Open(*wikiArchivePath)
		checkError(b, fileErr)

		ind := NewIndex()
		LoadWiki(archiveFile, func(a *Article) bool {
			ind.AddArticle(NewStrippedArticle(a))
			return true
		})
	})

	b.Run("LoadBzippedSync", func(b *testing.B) {
		bzipRaw, fileErr := os.Open(*wikiArchiveBzipPath)
		checkError(b, fileErr)
		bzipFile := bufio.NewReader(bzipRaw)
		archiveStream := bzip2.NewReader(bzipFile)

		ind := NewIndex()
		LoadWiki(archiveStream, func(a *Article) bool {
			ind.AddArticle(NewStrippedArticle(a))
			return true
		})
	})

	b.Run("LoadAsync", func(b *testing.B) {
		archiveFile, fileErr := os.Open(*wikiArchivePath)
		checkError(b, fileErr)
		ind := NewIndex()

		loadChan := make(chan *Article, 100)

		go func() {
			LoadWiki(archiveFile, func(a *Article) bool {
				loadChan <- a
				return true
			})
			close(loadChan)
		}()

		for a := range loadChan {
			ind.AddArticle(NewStrippedArticle(a))
		}

	})

	b.Run("LoadBzippedAsync", func(b *testing.B) {
		index, idxErr := os.Open(*WikiArchiveIndexBzipPath)
		archive, archiveErr := os.Open(*wikiArchiveBzipPath)
		checkError(b, idxErr)
		checkError(b, archiveErr)

		n := 0
		LoadWikiCompressed(index, archive, func(a *Article) bool {
			n++
			return true
		})

		if n < 100000 {
			b.Fatal("Did not load all articles, loaded", n)
		}

	})

}
