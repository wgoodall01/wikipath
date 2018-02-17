package wikipath

import (
	"os"
	"testing"
)

var A = &Article{Title: "A", Text: "[[B]] [[C]]"}
var B = &Article{Title: "B", Text: "[[A]] [[C]] [[D]]"}
var C = &Article{Title: "C", Text: "[[B]] [[E]]"}
var D = &Article{Title: "D", Text: ""}
var E = &Article{Title: "E", Redirect: Redirect{Title: "B"}}

func testRedirect(t *testing.T, index *Index) {
	redir := index.Get("E")
	if redir == nil {
		t.Fatal("Redirection: index.Get(\"E\") should return B, returned nil.")
	}
	if redir.Title != "B" {
		t.Fatal("Redirecting: index.Get(\"E\") should return B, returned", redir.Title)
	}
}

func testGet(t *testing.T, index *Index) {
	item := index.Get("B")
	if item == nil {
		t.Fatal("index.Get(\"B\") returned nil.")
	}
	if item.Title != "B" {
		t.Fatal("index.Get(\"B\") returned", item.Title)
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
		t.Log("Find path from A through D")
		path := index.FindPath(index.Get("A"), index.Get("D"), 20)
		t.Log(path)
		if path == nil {
			t.Fatal("Did not find path")
		}
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
		t.Log("Find path from B through D")
		path := index.FindPath(index.Get("B"), index.Get("D"), 20)
		t.Log(path)
		if path == nil {
			t.Fatal("Did not find path")
		}
	})
}

func die(b *testing.B, err error, msg string, args ...interface{}) {
	if err != nil {
		b.Logf(msg, args...)
		b.Fatal(err)
	}
}

func BenchmarkIndex(b *testing.B) {
	ind := NewIndex()

	b.Run("LoadWpindex", func(b *testing.B) {
		wpindexFile, wpindexFileErr := os.Open(*wpindexPath)
		die(b, wpindexFileErr, "Couldnt' open *.wpindex file.")

		wir, wirError := NewWpindexReader(wpindexFile)
		die(b, wirError, "Error creating wpindex reader.")

		for {
			article, readErr := wir.ReadArticle()
			if readErr == EOF {
				break
			} else if readErr != nil {
				die(b, readErr, "wpindex read error article=%v", article)
			} else {
				ind.AddArticle(article)
			}
		}
	})

	b.Run("BuildIndex", func(b *testing.B) {
		ind.Build()
	})
}
