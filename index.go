package main

import (
	"container/list"
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

func (path *IndexPath) ToSlice() []*IndexItem {
	pathArr := make([]*IndexItem, path.Len)
	for i := path.Len - 1; i >= 0; i-- {
		pathArr[i] = path.Item
		path = path.Prev
	}
	return pathArr
}

func (path *IndexPath) String() string {
	items := path.ToSlice()
	str := ""
	for i, it := range items {
		if i != 0 {
			str += " > "
		}
		str += it.Title
	}
	return str
}

type PathQueue struct {
	q *list.List
}

func NewPathQueue() *PathQueue {
	return &PathQueue{
		q: list.New(),
	}
}

func (pq *PathQueue) Enqueue(path *IndexPath) {
	pq.q.PushBack(path)
}

func (pq *PathQueue) Dequeue() *IndexPath {
	item := pq.q.Front()
	if item == nil {
		return nil
	}
	pq.q.Remove(item)
	return item.Value.(*IndexPath)
}

func NewIndex() *Index {
	return &Index{
		itemIndex:     make(map[string]*IndexItem),
		linkIndex:     make(map[string][]string),
		redirectIndex: make(map[string]string),
	}
}

// Get a list of paths between two IndexItems, sorted by length.
func (ind *Index) FindPath(from *IndexItem, to *IndexItem, depth int) *IndexPath {
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

	// Run the search.
	path := pathSearch(from, to, depth)

	return path
}

func pathSearch(from *IndexItem, to *IndexItem, depth int) *IndexPath {
	queue := NewPathQueue()
	rootPath := NewIndexPath(from)
	from.FoundPath = rootPath

	path := rootPath
	for ; path != nil; path = queue.Dequeue() {

	LinksLoop:
		for _, link := range path.Item.Links {
			linkPath := path.Append(link)
			if link.FoundPath != nil {
				// This item has already been searched.
				// Die.
				continue LinksLoop
			} else if link == to {
				// This is it, return the path.
				return linkPath
			} else {
				// If not, add to the queue of pages to be searched.
				queue.Enqueue(linkPath)
			}
		}
	}

	// No paths found.
	return nil
}

func (ind *Index) Reset() {
	if ind.dirty && ind.ready {
		for k, _ := range ind.itemIndex {
			ind.itemIndex[k].FoundPath = nil
		}
		ind.dirty = false
	}
}

func (ind *Index) AddArticle(a *StrippedArticle) {
	// Index these things:
	// - make an IndexItem, add it to the itemIndex
	// - Parse the links from the article text, add it to the linkIndex
	// - Figure out redirects, add them to the redirectIndex.
	k := NormalizeArticleTitle(a.Title)

	if a.Redirect != "" {
		// Item is a redirect, don't add it to the index.
		ind.redirectIndex[k] = NormalizeArticleTitle(a.Redirect)
	} else {
		// Item is normal, index it.
		ind.itemIndex[k] = &IndexItem{Title: a.Title}
		ind.linkIndex[k] = a.Links
	}

	ind.ready = false
}

func (ind *Index) Build() {
	if !ind.ready {
		// Go over indexed items
		for k, item := range ind.itemIndex {
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

		// Go over indexed redirects
		for k, redir := range ind.redirectIndex {
			redirItem, ok := ind.itemIndex[redir]

			// Only add link if item isn't a broken redirect
			if ok {
				ind.itemIndex[k] = redirItem
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
