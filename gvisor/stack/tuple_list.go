package stack

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
type tupleList struct {
	head *tuple
	tail *tuple
}
