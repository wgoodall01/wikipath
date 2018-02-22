package main

import (
	"encoding/json"
	wp "github.com/wgoodall01/wikipath/wp"
	"log"
	"net/http"
	"time"
)

type PathResponse struct {
	From     string   `json:"from"`     // Starting article
	To       string   `json:"to"`       // Ending article
	Path     []string `json:"path"`     // Path between articles.
	Duration float64  `json:"duration"` // Duration of query.
	Touched  int      `json:"touched"`  // How many articles touched.
}

type QueryHandler struct {
	ind *wp.Index
}

func NewQueryHandler(ind *wp.Index) *QueryHandler {
	return &QueryHandler{
		ind: ind,
	}
}

func (qh *QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fromName := query.Get("from")
	toName := query.Get("to")

	if fromName == "" || toName == "" {
		NewHttpError(http.StatusBadRequest, "Both 'from' and 'to' query parameters required").Send(w)
		return
	}

	// Get articles
	fromItem := qh.ind.Get(fromName)
	toItem := qh.ind.Get(toName)

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
	path, touched := qh.ind.FindPath(fromItem, toItem, MAX_DEPTH)
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
		Touched:  touched,
	}

	respBytes, respErr := json.MarshalIndent(resp, "", "  ")
	if respErr != nil {
		panic(respErr)
	}

	w.Write(respBytes)
	log.Printf("'%s' -> '%s' in %0.2f", fromName, toName, duration.Seconds())
}
