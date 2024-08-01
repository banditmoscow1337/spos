package buffer

import "io"

// Buffer is an alias to View.
type Buffer = View

// View is a non-linear buffer.
//
// All methods are thread compatible.
//
// +stateify savable
type View struct {
	data bufferList
	size int64
	pool pool
}

// VectorisedView is a vectorised version of View using non contiguous memory.
// It supports all the convenience methods supported by View.
//
// +stateify savable
type VectorisedView struct {
	views []View
	size  int
}

// NewVectorisedView creates a new vectorised view from an already-allocated
// slice of View and sets its size.
func NewVectorisedView(size int, views []View) VectorisedView {
	return VectorisedView{views: views, size: size}
}

// Read implements io.Reader.
func (vv *VectorisedView) Read(b []byte) (copied int, err error) {
	count := len(b)
	for count > 0 && len(vv.views) > 0 {
		if count < len(vv.views[0]) {
			vv.size -= count
			copy(b[copied:], vv.views[0][:count])
			vv.views[0].TrimFront(count)
			copied += count
			return copied, nil
		}
		count -= len(vv.views[0])
		copy(b[copied:], vv.views[0])
		copied += len(vv.views[0])
		vv.removeFirst()
	}
	if copied == 0 {
		return 0, io.EOF
	}
	return copied, nil
}
