package main

import (
	"encoding/xml"
	"io"
	"regexp"
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
		LoadWiki(source, func(a *Article) error {
			loadChan <- a
			return nil
		})
		close(loadChan)
	}()

	for a := range loadChan {
		ind.AddArticle(a)
	}

}

func LoadWiki(source io.Reader, visitor func(*Article) error) error {
	// Open an XML decoder over the file.
	decoder := xml.NewDecoder(source)

	for {
		// Get the next token.
		tok, tokError := decoder.Token()
		if tok == nil {
			break
		}
		if tokError != nil {
			return tokError
		}

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
