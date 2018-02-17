package wikipath

import (
	"runtime"
	"strings"
	"sync"
)

// NormalizeArticleTitle normalizes an article title to get a
// retrieval key
func NormalizeArticleTitle(title string) string {
	// Just lowercase it for now -- there are other normalization rules though.
	return strings.ToLower(title)
}

// Index contains all the loaded articles.
// If built, each article has a set of forward/reverse pointers.
type Index struct {
	itemIndex    map[string]*IndexItem // Map of normalized article title to `Item`s.
	itemIndexMut sync.RWMutex

	tempLinks  []*StrippedArticle // [empty if ready] Articles to be indexed.
	tempRedirs []*StrippedArticle // [empty if ready] Redirects to be indexed.

	dirty bool // If the items have been modified
	ready bool // If the index has been built
}

// IndexItem is an article in the index.
// If built, it has pointers to articles it links to, and pointers
// to articles which link to it.
type IndexItem struct {
	Title     string     // Non-normalized title of the page.
	FoundPath *IndexPath // If this item has been visited before.

	Forward    []*IndexItem // Items this pages links to.
	ForwardMut sync.Mutex

	Reverse    []*IndexItem // Items which link to this one.
	ReverseMut sync.Mutex
}

// NewIndex creates an Index.
func NewIndex() *Index {
	return &Index{
		itemIndex:  make(map[string]*IndexItem),
		tempLinks:  make([]*StrippedArticle, 0),
		tempRedirs: make([]*StrippedArticle, 0),
	}
}

// FindPath gets a list of paths between two IndexItems, sorted by length.
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

// Reset resets the index after a query.
// It goes through every article and deletes its FoundPath.
func (ind *Index) Reset() {
	if ind.dirty && ind.ready {
		for _, item := range ind.itemIndex {
			item.FoundPath = nil
		}
		ind.dirty = false
	}
}

// AddArticle adds an article to the index.
//
// Index these things:
// - make an IndexItem, add it to the itemIndex
// - Parse the links from the article text, add it to the linkIndex
// - Figure out redirects, add them to the redirectIndex.
func (ind *Index) AddArticle(a *StrippedArticle) {
	k := NormalizeArticleTitle(a.Title)

	if a.Redirect != "" {
		// Article is a redirect, add to redirect index.
		ind.tempRedirs = append(ind.tempRedirs, a)
	} else {
		// Article is not a redirect.

		// Make article if it doesn't already exist.
		if ind.itemIndex[k] == nil {
			ind.itemIndex[k] = &IndexItem{Title: a.Title}
		}

		// Add links to temp link index.
		ind.tempLinks = append(ind.tempLinks, a)
	}

	// Flag index as not ready.
	// Will rebuild on next use.
	ind.ready = false
}

// Build builds the index, finding each article's forward and reverse pointers.
func (ind *Index) Build() {

	itemPump := func(tempLinks []*StrippedArticle, tempItems chan<- *StrippedArticle) {
		for _, sa := range tempLinks {
			tempItems <- sa
		}
		close(tempItems)
	}

	linkWorker := func(ec *ErrorContext, tempItems <-chan *StrippedArticle) {
		// read tempItem from chan
		for {
			sa := <-tempItems
			if sa == nil {
				break
			}

			k := NormalizeArticleTitle(sa.Title)

			// tempItem isn't a redirect

			// For each link in the article, add a forward and reverse pointer
			linkSrc := ind.Get(k) // Get() takes care of locking
			for _, linkName := range sa.Links {
				linkDst := ind.Get(linkName)

				// Check for broken links.
				if linkDst != nil {
					// Append to forward links
					linkSrc.ForwardMut.Lock()
					linkSrc.Forward = append(linkSrc.Forward, linkDst)
					linkSrc.ForwardMut.Unlock()

					// Append to reverse links
					linkDst.ReverseMut.Lock()
					linkDst.Reverse = append(linkDst.Reverse, linkSrc)
					linkDst.ReverseMut.Unlock()
				}
			}
		}
		ec.Done()
	}

	linksWait := NewErrorContext()

	// Channel of all the temp items which need indexing.
	tempItems := make(chan *StrippedArticle)
	go itemPump(ind.tempLinks, tempItems)

	nWorkers := runtime.GOMAXPROCS(-1)
	for i := 0; i < nWorkers; i++ {
		linksWait.Add(1)
		go linkWorker(linksWait, tempItems)
	}

	// Wait for all link workers to finish indexing.
	linksWait.Wait()

	// Index all redirects.
	for _, sa := range ind.tempRedirs {
		k := NormalizeArticleTitle(sa.Title)
		redir := ind.Get(sa.Redirect)

		// Check for broken links.
		if redir != nil {
			ind.itemIndex[k] = redir
		}
	}

	// Remove temp index, it's unneeded.
	ind.tempLinks = nil
	ind.tempRedirs = nil

	// Index is now ready.
	ind.ready = true
}

// Status gets status of index as (ready, dirty)
func (ind *Index) Status() (ready bool, dirty bool) {
	return ind.ready, ind.dirty
}

// Get gets an IndexItem by article title.
func (ind *Index) Get(title string) *IndexItem {
	k := NormalizeArticleTitle(title)
	ind.itemIndexMut.RLock()
	defer ind.itemIndexMut.RUnlock()
	return ind.itemIndex[k]
}
