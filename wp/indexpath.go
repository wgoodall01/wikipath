package wikipath

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
