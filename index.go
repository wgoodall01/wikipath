package main

import (
	"sort"
	"strings"
)

func NormalizeArticleTitle(title string) string {
	// Just lowercase it for now -- there are other normalization rules though.
	return strings.ToLower(title)
}

type Index struct {
	itemIndex     map[string]*IndexItem // Map of normalized article title to `Item`s.
	linkIndex     map[string][]string   // [nil if ready] Map of norm. titles to their normalized links.
	redirectIndex map[string]string     // [nil if ready] Map of norm. titles to norm. titles representing redirects.
	dirty         bool                  // If the items have been modified
	ready         bool                  // If the index has been built
}

type IndexItem struct {
	Title     string       // Non-normalized title of the page.
	FoundPath *IndexPath   // If this item has been visited before.
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

func NewIndex() *Index {
	return &Index{
		itemIndex:     make(map[string]*IndexItem),
		linkIndex:     make(map[string][]string),
		redirectIndex: make(map[string]string),
	}
}

// Get a list of paths between two IndexItems, sorted by length.
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

	// Sort paths by length.
	sort.Slice(pathList, func(i int, j int) bool {
		return len(pathList[i]) < len(pathList[j])
	})

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
	// Index these things:
	// - make an IndexItem, add it to the itemIndex
	// - Parse the links from the article text, add it to the linkIndex
	// - Figure out redirects, add them to the redirectIndex.

	k := NormalizeArticleTitle(a.Title)
	ind.itemIndex[k] = &IndexItem{Title: a.Title}
	ind.linkIndex[k] = ParseLinks(a.Text)

	if a.Redirect.Title != "" {
		ind.redirectIndex[k] = NormalizeArticleTitle(a.Redirect.Title)
	}

	ind.ready = false
}

func (ind *Index) Build() {
	if !ind.ready {
		for k, item := range ind.itemIndex {
			redir := ind.redirectIndex[k]
			if redir != "" {
				// Item is a redirect, add pointer to next article.
				ind.itemIndex[k] = ind.itemIndex[redir]
			} else {
				articleLinks := ind.linkIndex[k]
				item.Links = make([]*IndexItem, 0, len(articleLinks))

				for _, linkName := range articleLinks {
					linkItem := ind.Get(linkName)

					// Check for broken links. Wikipedia isn't perfect.
					if linkItem != nil {
						item.Links = append(item.Links, linkItem)
					}
				}
			}
		}
	}

	// Remove link, redirect indexes, they're unneeded.
	// Redirects are now built in to the itemIndex
	ind.redirectIndex = nil
	ind.linkIndex = nil

	// Index is now ready.
	ind.ready = true
}

// Get status of index as (ready, dirty)
func (ind *Index) Status() (ready bool, dirty bool) {
	return ind.ready, ind.dirty
}

// Get an IndexItem by article title.
// Follows any and all redirects.
func (ind *Index) Get(title string) *IndexItem {
	k := NormalizeArticleTitle(title)
	redir := ind.redirectIndex[k] // Get a redirect
	if redir != "" {
		// Follow ONE redirect to an article.
		// Wikipedia doesn't allow for >1 redirect, so there can't be loops.
		return ind.itemIndex[redir]
	} else {
		// Return the item from the index
		return ind.itemIndex[k]
	}
}
