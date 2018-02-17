package wikipath

import (
	"compress/gzip"
	"encoding/gob"
	"io"
)

const compressionLevel = gzip.BestSpeed

var EOF error = io.EOF

type StrippedArticle struct {
	Title    string
	Id       int
	Redirect string
	Links    []string
}

func NewStrippedArticle(a *Article) *StrippedArticle {
	return &StrippedArticle{
		Title:    a.Title,
		Redirect: a.Redirect.Title,
		Links:    ParseLinks(a.Text),
		Id:       a.Id,
	}
}

type WpindexWriter struct {
	writer     io.Writer
	gzipWriter *gzip.Writer
	gobEncoder *gob.Encoder
}

func NewWpindexWriter(f io.Writer) *WpindexWriter {
	gzipWriter, _ := gzip.NewWriterLevel(f, compressionLevel)
	gobEncoder := gob.NewEncoder(gzipWriter)
	return &WpindexWriter{
		writer:     f,
		gzipWriter: gzipWriter,
		gobEncoder: gobEncoder,
	}
}

func (this *WpindexWriter) WriteArticle(a *StrippedArticle) error {
	return this.gobEncoder.Encode(a)
}

func (this *WpindexWriter) Close() error {
	gzipErr := this.gzipWriter.Close()
	return gzipErr
}

type WpindexReader struct {
	reader     io.Reader
	gzipReader *gzip.Reader
	gobDecoder *gob.Decoder
}

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

func (this *WpindexReader) ReadArticle() (*StrippedArticle, error) {
	var a StrippedArticle
	err := this.gobDecoder.Decode(&a)
	return &a, err
}

func (this *WpindexReader) Close() error {
	return this.gzipReader.Close()
}
