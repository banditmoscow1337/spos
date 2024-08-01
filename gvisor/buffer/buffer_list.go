package buffer

// List is an intrusive list. Entries can be added to or removed from the list
// in O(1) time and with no additional memory allocations.
//
// The zero value for List is an empty list ready to use.
//
// To iterate over a list (where l is a List):
//
//	     for e := l.Front(); e != nil; e = e.Next() {
//			// do something with e.
//	     }
//
// +stateify savable
type bufferList struct {
	head *buffer
	tail *buffer
}

// Entry is a default implementation of Linker. Users can add anonymous fields
// of this type to their structs to make them automatically implement the
// methods needed by List.
//
// +stateify savable
type bufferEntry struct {
	next *buffer
	prev *buffer
}
