package main

import "testing"

var A = &Article{Title: "A", Text: "[[B]]"}
var B = &Article{Title: "B", Text: "[[C]]"}
var C = &Article{Title: "C", Text: "[[D]] [[A]]"}
var D = &Article{Title: "D", Text: "[[A]]"}

func TestIndex(t *testing.T) {
	index := NewIndex()
	for _, article := range []*Article{A, B, C, D} {
		index.AddArticle(article)
	}

	index.MakeIndex()

	if index.Get("A").Article != A {
		t.Logf("index.Get() didn't return right article.")
	}

	FindPath(index.Get("B"), index.Get("A"), 4)
}
