package main

import "strings"

func NormalizeArticleTitle(title string) string {
	// Just lowercase it for now -- there are other normalization rules though.
	return strings.ToLower(title)
}

type Index struct {
	itemIndex map[string]*IndexItem // Map of normalized article title to `Item`s.
}

type IndexItem struct {
	Title   string       // Title of the page.
	Visited bool         // If this item has been visited before.
	Article *Article     // Pointer to that Item's article.
	Links   []*IndexItem // Pointers to other items, representing the links on the page.
}

type IndexPath struct {
	Item *IndexItem // Item in the path
	Prev *IndexPath // The previous item, `nil` if the starting item.
}

func FindPath(from *IndexItem, to *IndexItem, depth int8) [][]*IndexItem {
	rootPath := IndexPath{Item: from, Prev: nil}
	validPaths := rootPath.PathsTo(to, depth)

	pathList := make([][]*IndexItem, len(validPaths))
	for i, head := range validPaths {
		orig := head.Prev
		pathLen := 1
		for orig.Prev != nil {
			orig = orig.Prev
			pathLen += 1
		}

		path = make([]*IndexItem, pathLen+1)
		for j := pathLen; j > 0; j-- {
			path[j] = head.Item
			head = head.Prev
		}

		pathList[i] = path
	}
}

// Returns a list of valid `IndexPath`s from `path` to `to`.
func (path *IndexPath) PathsTo(to *IndexItem, depth int8) []IndexPath {
	validPaths := make([]IndexPath, 0)
	if depth > 0 {
		for _, link := range path.Item.Links {
			linkPath := IndexPath{Item: link, Prev: path}
			if link.Visited {
				// Ignore visited links.
			} else if link == to {
				// If the link is to the final location, append the link's path.
				validPaths = append(validPaths, linkPath)
			} else {
				// If not, search the sub-links and append those paths.
				validPaths = append(validPaths, linkPath.PathsTo(to, depth-1)...)
			}
		}
	}
	return validPaths
}

func NewIndex() Index {
	return Index{
		itemIndex: make(map[string]*IndexItem),
	}
}

func (ind *Index) AddArticle(a *Article) {
	// Add it to the article index.
	k := NormalizeArticleTitle(a.Title)
	ind.itemIndex[k] = &IndexItem{Title: a.Title, Article: a}
}

func (ind *Index) MakeIndex() {
	for _, item := range ind.itemIndex {
		// Create an 'Item' for it, add that to the index.
		articleLinks := ParseLinks(item.Article.Text)
		item.Links = make([]*IndexItem, len(articleLinks))

		for i, link := range articleLinks {
			item.Links[i] = ind.itemIndex[NormalizeArticleTitle(link)]
		}
	}
}

func (ind *Index) Get(title string) *IndexItem {
	k := NormalizeArticleTitle(title)
	return ind.itemIndex[k]
}
