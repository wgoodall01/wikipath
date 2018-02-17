package wikipath

import (
	"compress/gzip"
	"encoding/gob"
	"io"
)

const compressionLevel = gzip.BestSpeed

// EOF is the error returned when the `*.wpindex` file ends.
var EOF = io.EOF

// StrippedArticle is an article, stripped of everything save for its
// title, id, redirect title, and string links.
type StrippedArticle struct {
	Title    string
	ID       int
	Redirect string
	Links    []string
}

// NewStrippedArticle creates a StrippedArticle from an Article.
func NewStrippedArticle(a *Article) *StrippedArticle {
	return &StrippedArticle{
		Title:    a.Title,
		Redirect: a.Redirect.Title,
		Links:    ParseLinks(a.Text),
		ID:       a.ID,
	}
}

// WpindexWriter writes `StrippedArticle`s to a *.wpindex file.
type WpindexWriter struct {
	writer     io.Writer
	gzipWriter *gzip.Writer
	gobEncoder *gob.Encoder
}

// NewWpindexWriter creates a `WpindexWriter`.
func NewWpindexWriter(f io.Writer) *WpindexWriter {
	gzipWriter, _ := gzip.NewWriterLevel(f, compressionLevel)
	gobEncoder := gob.NewEncoder(gzipWriter)
	return &WpindexWriter{
		writer:     f,
		gzipWriter: gzipWriter,
		gobEncoder: gobEncoder,
	}
}

// WriteArticle writes an article to the *.wpindex file.
func (wiw *WpindexWriter) WriteArticle(a *StrippedArticle) error {
	return wiw.gobEncoder.Encode(a)
}

// Close closes the `WpindexWriter`.
func (wiw *WpindexWriter) Close() error {
	gzipErr := wiw.gzipWriter.Close()
	return gzipErr
}

// WpindexReader reads articles from a *.wpindex file.
type WpindexReader struct {
	reader     io.Reader
	gzipReader *gzip.Reader
	gobDecoder *gob.Decoder
}

// NewWpindexReader creates a `WpindexReader` from an `io.Reader`.
func NewWpindexReader(f io.Reader) (*WpindexReader, error) {
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	gobDecoder := gob.NewDecoder(gzipReader)
	return &WpindexReader{
		reader:     f,
		gzipReader: gzipReader,
		gobDecoder: gobDecoder,
	}, nil
}

// ReadArticle reads an article from the `WpindexReader`
func (wir *WpindexReader) ReadArticle() (*StrippedArticle, error) {
	var a StrippedArticle
	err := wir.gobDecoder.Decode(&a)
	return &a, err
}

// Close closes the `WpindexReader`.
func (wir *WpindexReader) Close() error {
	return wir.gzipReader.Close()
}
