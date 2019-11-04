package main

// Generate the static bundle.
//go:generate go run github.com/rakyll/statik -src=./wikipath-web/build/

import (
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

	http.Handle("/api/query", NewQueryHandler(idx))
	http.Handle("/api/random", NewRandomHandler(idx))
	http.Handle("/", http.FileServer(statikFS))

	log.Printf("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
