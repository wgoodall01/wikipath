package main

import (
	"encoding/json"
	wp "github.com/wgoodall01/wikipath/wp"
	"log"
	"net/http"
)

type RandomResponse struct {
	Title string `json:"title"` // Title of the article.
}

type RandomHandler struct {
	ind *wp.Index
}

func NewRandomHandler(ind *wp.Index) *RandomHandler {
	return &RandomHandler{
		ind: ind,
	}
}

func (rh *RandomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	article := rh.ind.GetRandom()
	bytes, respErr := json.MarshalIndent(RandomResponse{Title: article.Title}, "", "  ")
	if respErr != nil {
		panic(respErr)
	}
	w.Write(bytes)
	log.Printf("Random article '%s'", article.Title)
}
