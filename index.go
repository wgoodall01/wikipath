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
	Forward   []*IndexItem // Items this pages links to.
	Reverse   []*IndexItem // Items which link to this one.
}

type Direction bool

const FORWARD Direction = true
const REVERSE Direction = false

type IndexPath struct {
	Item      *IndexItem // Item in the path
	Prev      *IndexPath // The previous item, `nil` if the starting item.
	Len       int        // Length of the path
	Direction Direction  // The direction the path is from.
}

func NewIndexPath(it *IndexItem, direction Direction) *IndexPath {
	return &IndexPath{
		Item:      it,
		Prev:      nil,
		Len:       1,
		Direction: direction,
	}
}

// Join a forward and a reverse path together into one. The
// heads of each path must point to the same item.
// Returns nil if it doesn't work out.
func NewIndexPathByJoin(i1 *IndexPath, i2 *IndexPath) *IndexPath {
	var all *IndexPath
	var rest *IndexPath

	if i1.Direction == FORWARD && i2.Direction == REVERSE {
		all = i1
		rest = i2
	} else if i1.Direction == REVERSE && i2.Direction == FORWARD {
		all = i2
		rest = i1
	} else {
		// They have to be one forwards, one reverse.
		return nil
	}

	if rest.Item != all.Item {
		// They don't join up evenly. Die.
		return nil
	}

	// Advance rest by 1, skip
	rest = rest.Prev

	for rest != nil {
		all = all.Append(rest.Item)
		rest = rest.Prev
	}

	return all
}

func (path *IndexPath) Append(it *IndexItem) *IndexPath {
	return &IndexPath{
		Item:      it,
		Prev:      path,
		Len:       path.Len + 1,
		Direction: path.Direction,
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
	// Idiot check
	if from == to {
		return NewIndexPath(from, FORWARD)
	}

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

	fromPath := NewIndexPath(from, FORWARD)
	toPath := NewIndexPath(to, REVERSE)
	queue.Enqueue(toPath)

	path := fromPath

	for ; path != nil; path = queue.Dequeue() {

		// Get the right link list depending on direction
		var links []*IndexItem
		if path.Direction == FORWARD {
			links = path.Item.Forward
		} else {
			links = path.Item.Reverse
		}

		for _, link := range links {
			linkPath := path.Append(link)

			if (linkPath.Direction == FORWARD && link == to) || (linkPath.Direction == REVERSE && link == from) {
				// Found the end of the path.
				return linkPath
			} else if link.FoundPath == nil {
				// Not searched yet. Add this to the queue of pages to be searched.
				link.FoundPath = linkPath
				queue.Enqueue(linkPath)
			} else if link.FoundPath.Direction == linkPath.Direction {
				// Already searched in the same direction.
				// Ignore it.
			} else {
				// At this point, link.FoundPath.Direction != linkPath.Direction
				// We've met the search coming from the other direction.
				return NewIndexPathByJoin(linkPath, link.FoundPath)
			}
		}
	}

	// Nothing happened.
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
			item.Forward = make([]*IndexItem, 0, len(articleLinks))

			for _, linkName := range articleLinks {
				linkItem := ind.Get(linkName)

				// Check for broken links. Wikipedia isn't perfect.
				if linkItem != nil {
					// Append to forward and reverse link sets.
					item.Forward = append(item.Forward, linkItem)
					linkItem.Reverse = append(linkItem.Reverse, item)
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
