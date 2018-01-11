package main

import "strings"

func NormalizeArticleTitle(title string) string {
	// Just lowercase it for now -- there are other normalization rules though.
	return strings.ToLower(title)
}

type Index struct {
	itemIndex map[string]*IndexItem // Map of normalized article title to `Item`s.
	dirty     bool                  // If the items have been modified
	ready     bool                  // If the index has been built
}

type IndexItem struct {
	Title     string       // Title of the page.
	FoundPath *IndexPath   // If this item has been visited before.
	Article   *Article     // Pointer to that Item's article.
	Links     []*IndexItem // Pointers to other items, representing the links on the page.
}

type IndexPath struct {
	Item *IndexItem // Item in the path
	Prev *IndexPath // The previous item, `nil` if the starting item.
	Len  int        // Length of the path
}

func NewIndexPath(it *IndexItem) *IndexPath {
	return &IndexPath{
		Item: it,
		Prev: nil,
		Len:  1,
	}
}

func (path *IndexPath) Append(it *IndexItem) *IndexPath {
	return &IndexPath{
		Item: it,
		Prev: path,
		Len:  path.Len + 1,
	}
}

// Returns a list of valid `IndexPath`s from `path` to `to`.
func (path *IndexPath) PathsTo(to *IndexItem, depth int8) []*IndexPath {
	validPaths := make([]*IndexPath, 0)
	if depth > 0 {
		for _, link := range path.Item.Links {
			linkPath := path.Append(link)
			if link.FoundPath != nil && link.FoundPath.Len < path.Len {
				// Ignore links already visited with a shorter path.
				// Die.
				return validPaths
			} else if link == to {
				// If the link is to the final location, append the link's path.
				validPaths = append(validPaths, linkPath)
			} else {
				// If not, search the sub-links and append those paths.
				validPaths = append(validPaths, linkPath.PathsTo(to, depth-1)...)
			}

			// Set found path.
			path.Item.FoundPath = path
		}
	}

	return validPaths
}

func NewIndex() Index {
	return Index{
		itemIndex: make(map[string]*IndexItem),
	}
}

func (ind *Index) FindPath(from *IndexItem, to *IndexItem, depth int8) [][]*IndexItem {
	// Ensure index is clean.
	if ind.dirty {
		ind.Reset()
	} else {
		ind.dirty = true
	}

	// Ensure index has been built.
	if !ind.ready {
		ind.Build()
	}

	rootPath := NewIndexPath(from)
	validPaths := rootPath.PathsTo(to, depth)

	pathList := make([][]*IndexItem, len(validPaths))
	for i, head := range validPaths {
		path := make([]*IndexItem, head.Len)
		for j := head.Len - 1; j >= 0; j-- {
			path[j] = head.Item
			head = head.Prev
		}

		pathList[i] = path
	}

	return pathList
}

func (ind *Index) Reset() {
	if ind.dirty && ind.ready {
		for _, it := range ind.itemIndex {
			it.FoundPath = nil
		}
		ind.dirty = false
	}
}

func (ind *Index) AddArticle(a *Article) {
	// Add it to the article index.
	k := NormalizeArticleTitle(a.Title)
	ind.itemIndex[k] = &IndexItem{Title: a.Title, Article: a}
	ind.ready = false
}

func (ind *Index) Build() {
	if !ind.ready {
		for _, item := range ind.itemIndex {
			// Create an 'Item' for it, add that to the index.
			articleLinks := ParseLinks(item.Article.Text)
			item.Links = make([]*IndexItem, len(articleLinks))

			for i, link := range articleLinks {
				item.Links[i] = ind.itemIndex[NormalizeArticleTitle(link)]
			}
		}
	}
	ind.ready = true
}

// Get status of index as (ready, dirty)
func (ind *Index) Status() (ready bool, dirty bool) {
	return ind.ready, ind.dirty
}

func (ind *Index) Get(title string) *IndexItem {
	k := NormalizeArticleTitle(title)
	return ind.itemIndex[k]
}
