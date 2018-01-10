package main

import "strings"

type Index struct {
	itemIndex map[string]*Item // Map of normalized article title to `Item`s.
}

type Item struct {
	Title   string   // Title of the page.
	Article *Article // Pointer to that Item's article.
	Links   []*Item  // Pointers to other items, representing the links on the page.
}

func NormalizeArticleTitle(title string) string {
	// Just lowercase it for now -- there are other normalization rules though.
	return strings.ToLower(title)
}

func NewIndex() Index {
	return Index{
		itemIndex: make(map[string]*Item),
	}
}

func (ind *Index) AddArticle(a Article) {
	// Add it to the article index.
	k := NormalizeArticleTitle(a.Title)
	ind.itemIndex[k] = &Item{Title: a.Title, Article: &a}
}

func (ind *Index) MakeIndex() {
	for _, item := range ind.itemIndex {
		// Create an 'Item' for it, add that to the index.
		articleLinks := ParseLinks(item.Article.Text)
		item.Links = make([]*Item, len(articleLinks))

		for i, link := range articleLinks {
			item.Links[i] = ind.itemIndex[NormalizeArticleTitle(link)]
		}
	}
}

func (ind *Index) Get(title string) Item {
	k := NormalizeArticleTitle(title)
	return *ind.itemIndex[k]
}
