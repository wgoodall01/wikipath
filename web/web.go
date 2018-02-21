package main

// Generate the static bundle.
//go:generate statik -src=./wikipath-web/build/

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	// Wikipath
	wp "github.com/wgoodall01/wikipath/wp"

	// Static files
	"github.com/rakyll/statik/fs"
	_ "github.com/wgoodall01/wikipath/web/statik"
)

const MAX_DEPTH int = 10 // Maximum query depth.

type PathResponse struct {
	From     string   `json:"from"`     // Starting article
	To       string   `json:"to"`       // Ending article
	Path     []string `json:"path"`     // Path between articles.
	Duration float64  `json:"duration"` // Duration of query.
}

type HttpError struct {
	Status  int    `json:"status"` // Status code
	kind    string // Unique kind
	Message string `json:"message"` // Descriptive message
}

func NewHttpError(status int, message string) *HttpError {
	return &HttpError{
		Status:  status,
		Message: message,
	}
}

func (he *HttpError) Send(w http.ResponseWriter) {
	w.WriteHeader(he.Status)
	bytes, _ := json.MarshalIndent(he, "", "  ")
	w.Write(bytes)
}

func main() {
	log.Printf(" -- Starting Wikipath -- ")

	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("fatal: expected 1 argument, got %d", len(args))
	}
	indexPath := os.Args[1]

	indexFile, fileErr := os.Open(indexPath)
	if fileErr != nil {
		log.Fatalf("fatal: couldn't open index: %v", fileErr)
	}

	wir, wirErr := wp.NewWpindexReader(indexFile)
	if wirErr != nil {
		log.Fatalf("fatal: couldn't create wpindex reader: %v", wirErr)
	}

	log.Printf("Loading index from '%s'...", indexPath)

	// Load articles to index
	startLoad := time.Now()
	idx := wp.NewIndex()

	for {
		sa, readErr := wir.ReadArticle()
		if readErr == wp.EOF {
			break // end of file
		} else if readErr != nil {
			log.Fatalf("fatal: error loading from index: %v", readErr)
		} else {
			idx.AddArticle(sa)
		}
	}

	durLoad := time.Since(startLoad)
	log.Printf("Loaded index in %.2fs", durLoad.Seconds())

	// Build index.
	log.Printf("Building index...")
	startBuild := time.Now()
	idx.Build()
	durBuild := time.Since(startBuild)
	log.Printf("Built index in %.2fs", durBuild.Seconds())

	// Start webserver.
	statikFS, statikErr := fs.New()
	if statikErr != nil {
		log.Fatalf("err: statik: %v", statikErr)
	}
	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		fromName := query.Get("from")
		toName := query.Get("to")

		if fromName == "" || toName == "" {
			NewHttpError(http.StatusBadRequest, "Both 'from' and 'to' query parameters required").Send(w)
			return
		}

		// Get articles
		fromItem := idx.Get(fromName)
		toItem := idx.Get(toName)

		if fromItem == nil {
			NewHttpError(http.StatusNotFound, "Could not find 'from' article.").Send(w)
			return
		}
		if toItem == nil {
			NewHttpError(http.StatusNotFound, "Could not find 'to' article").Send(w)
			return
		}

		// Find path.
		tStart := time.Now()
		path := idx.FindPath(fromItem, toItem, MAX_DEPTH)
		duration := time.Since(tStart)

		if path == nil {
			NewHttpError(http.StatusNotFound, "Could not find valid path").Send(w)
			return
		}

		titles := path.ToStringSlice()
		resp := PathResponse{
			From:     fromName,
			To:       toName,
			Path:     titles,
			Duration: duration.Seconds(),
		}

		respBytes, respErr := json.MarshalIndent(resp, "", "  ")
		if respErr != nil {
			panic(respErr)
		}

		w.Write(respBytes)
		log.Printf("'%s' -> '%s' in %0.2f", fromName, toName, duration.Seconds())
	})
	http.Handle("/", http.FileServer(statikFS))

	log.Printf("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
