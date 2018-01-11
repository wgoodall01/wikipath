package main

import (
	"testing"
)

var A = &Article{Title: "A", Text: "[[B]] [[C]]"}
var B = &Article{Title: "B", Text: "[[C]] [[D]]"}
var C = &Article{Title: "C", Text: "[[B]]"}
var D = &Article{Title: "D", Text: ""}

func TestIndex(t *testing.T) {
	index := NewIndex()

	ready0, dirty0 := index.Status()
	t.Logf("Created index. ready:%t dirty:%t", ready0, dirty0)

	for _, article := range []*Article{A, B, C, D} {
		index.AddArticle(article)
	}

	ready1, dirty1 := index.Status()
	t.Logf("Loaded articles. ready:%t dirty:%t", ready1, dirty1)

	index.Build()

	ready2, dirty2 := index.Status()
	t.Logf("Built index. ready:%t dirty:%t", ready2, dirty2)

	if index.Get("A").Article != A {
		t.Logf("index.Get() didn't return right article.")
	}

	test := index.FindPath(index.Get("A"), index.Get("D"), 20)

	ready3, dirty3 := index.Status()
	t.Logf("Paths found. ready:%t dirty:%t", ready3, dirty3)

	for _, path := range test {
		str := ""
		for _, item := range path {
			str = str + item.Title + ", "
		}
		t.Log("Path: " + str)
	}
}
