package main

import (
	"testing"

	_ "github.com/davecgh/go-spew/spew"
)

var A = &Article{Title: "A", Text: "[[B]] [[C]]"}
var B = &Article{Title: "B", Text: "[[A]] [[C]] [[D]]"}
var C = &Article{Title: "C", Text: "[[B]] [[E]]"}
var D = &Article{Title: "D", Text: ""}
var E = &Article{Title: "E", Redirect: Redirect{Title: "B"}}

func testRedirect(t *testing.T, index *Index) {
	redir := index.Get("E")
	if redir.Title != "B" {
		t.Fatal("Redirecting: index.Get(\"E\") should return B, returned", redir.Title)
	}
}

func testGet(t *testing.T, index *Index) {
	title := index.Get("B").Title
	if title != "B" {
		t.Fatal("index.Get(\"B\") returned", title)
	}
}

func TestIndex(t *testing.T) {
	index := NewIndex()

	t.Run("Status0", func(t *testing.T) {
		ready, dirty := index.Status()
		if ready != false {
			t.Fatal("Reports ready before Build() is called")
		}
		if dirty != false {
			t.Fatal("Reports dirty before FindPath() is called")
		}
	})

	t.Run("AddArticles", func(t *testing.T) {
		for _, article := range []*Article{A, B, C, D, E} {
			index.AddArticle(NewStrippedArticle(article))
		}
	})

	t.Run("Status1", func(t *testing.T) {
		ready, dirty := index.Status()
		if ready != false {
			t.Fatal("Reports ready before Build() is called")
		}
		if dirty != false {
			t.Fatal("Reports dirty before FindPath() is called")
		}
	})

	t.Run("AccessPreBuild", func(t *testing.T) {
		testRedirect(t, index)
		testGet(t, index)
	})

	t.Run("Build", func(t *testing.T) {
		index.Build()
	})

	t.Run("Status2", func(t *testing.T) {
		ready, dirty := index.Status()
		if ready != true {
			t.Fatal("Reports not ready after Build() is called")
		}
		if dirty != false {
			t.Fatal("Reports dirty before FindPath() is called")
		}
	})

	t.Run("AccessPostBuild", func(t *testing.T) {
		testGet(t, index)
		testRedirect(t, index)
	})

	t.Run("PathFind1", func(t *testing.T) {
		path := index.FindPath(index.Get("A"), index.Get("D"), 20)
		t.Log(path)
	})

	t.Run("Status3", func(t *testing.T) {
		ready, dirty := index.Status()
		if ready != true {
			t.Fatal("Reports Not ready after Build() is called")
		}
		if dirty != true {
			t.Fatal("Reports not dirty after pathfind")
		}
	})

	t.Run("PathFind2", func(t *testing.T) {
		path := index.FindPath(index.Get("B"), index.Get("D"), 20)
		t.Log(path)
	})
}
