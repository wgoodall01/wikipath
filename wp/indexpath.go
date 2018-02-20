package wikipath

// Direction is either forward or reverse. This represents
// a direction for the bidirectional search.
type Direction bool

// FORWARD goes forward
const FORWARD Direction = true

// REVERSE goes in reverse
const REVERSE Direction = false

// IndexPath represents a link path through the Index.
type IndexPath struct {
	Item      *IndexItem // Item in the path
	Prev      *IndexPath // The previous item, `nil` if the starting item.
	Direction Direction  // The direction the path is from.
}

// NewIndexPath creates an IndexPath from a starting node and a Direction.
func NewIndexPath(it *IndexItem, direction Direction) *IndexPath {
	return &IndexPath{
		Item:      it,
		Prev:      nil,
		Direction: direction,
	}
}

// NewIndexPathByJoin joins a forward and a reverse path together into one. The
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

// Append returns a new IndexPath by appending `it` to `path`.
func (path *IndexPath) Append(it *IndexItem) *IndexPath {
	return &IndexPath{
		Item:      it,
		Prev:      path,
		Direction: path.Direction,
	}
}

// ToSlice returns the IndexPath as a []*IndexItem.
func (path *IndexPath) ToSlice() []*IndexItem {
	len := path.Len()
	pathArr := make([]*IndexItem, len)
	for i := len - 1; i >= 0; i-- {
		pathArr[i] = path.Item
		path = path.Prev
	}
	return pathArr
}

func (path *IndexPath) Len() int {
	n := 1
	itm := path
	for itm.Prev != nil {
		itm = itm.Prev
		n++
	}
	return n
}

// String converts the IndexPath to a string.
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
