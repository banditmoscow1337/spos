package waiter

import "sync"

// Queue represents the wait queue where waiters can be added and
// notifiers can notify them when events happen.
//
// The zero value for waiter.Queue is an empty queue ready for use.
//
// +stateify savable
type Queue struct {
	list waiterList
	mu   sync.RWMutex `state:"nosave"`
}

// EventMask represents io events as used in the poll() syscall.
type EventMask uint64

// EntryCallback provides a notify callback.
type EntryCallback interface {
	// Callback is the function to be called when the waiter entry is
	// notified. It is responsible for doing whatever is needed to wake up
	// the waiter.
	//
	// The callback is supposed to perform minimal work, and cannot call
	// any method on the queue itself because it will be locked while the
	// callback is running.
	//
	// The mask indicates the events that occurred and that the entry is
	// interested in.
	Callback(e *Entry, mask EventMask)
}

// Events that waiters can wait on. The meaning is the same as those in the
// poll() syscall.
const (
	EventIn     EventMask = 0x01   // POLLIN
	EventPri    EventMask = 0x02   // POLLPRI
	EventOut    EventMask = 0x04   // POLLOUT
	EventErr    EventMask = 0x08   // POLLERR
	EventHUp    EventMask = 0x10   // POLLHUP
	EventRdNorm EventMask = 0x0040 // POLLRDNORM
	EventWrNorm EventMask = 0x0100 // POLLWRNORM

	allEvents      EventMask = 0x1f | EventRdNorm | EventWrNorm
	ReadableEvents EventMask = EventIn | EventRdNorm
	WritableEvents EventMask = EventOut | EventWrNorm
)

// ToLinux returns e in the format used by Linux poll(2).
func (e EventMask) ToLinux() uint32 {
	// Our flag definitions are currently identical to Linux.
	return uint32(e)
}

// Entry represents a waiter that can be add to the a wait queue. It can
// only be in one queue at a time, and is added "intrusively" to the queue with
// no extra memory allocations.
//
// +stateify savable
type Entry struct {
	Callback EntryCallback

	// The following fields are protected by the queue lock.
	mask EventMask
	waiterEntry
}

// Entry is a default implementation of Linker. Users can add anonymous fields
// of this type to their structs to make them automatically implement the
// methods needed by List.
//
// +stateify savable
type waiterEntry struct {
	next *Entry
	prev *Entry
}

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
type waiterList struct {
	head *Entry
	tail *Entry
}

// ElementMapper provides an identity mapping by default.
//
// This can be replaced to provide a struct that maps elements to linker
// objects, if they are not the same. An ElementMapper is not typically
// required if: Linker is left as is, Element is left as is, or Linker and
// Element are the same type.
type waiterElementMapper struct{}

// linkerFor maps an Element to a Linker.
//
// This default implementation should be inlined.
//
//go:nosplit
func (waiterElementMapper) linkerFor(elem *Entry) *Entry { return elem }

// PushBack inserts the element e at the back of list l.
//
//go:nosplit
func (l *waiterList) PushBack(e *Entry) {
	linker := waiterElementMapper{}.linkerFor(e)
	linker.SetNext(nil)
	linker.SetPrev(l.tail)
	if l.tail != nil {
		waiterElementMapper{}.linkerFor(l.tail).SetNext(e)
	} else {
		l.head = e
	}

	l.tail = e
}

// SetNext assigns 'entry' as the entry that follows e in the list.
//
//go:nosplit
func (e *waiterEntry) SetNext(elem *Entry) {
	e.next = elem
}

// SetPrev assigns 'entry' as the entry that precedes e in the list.
//
//go:nosplit
func (e *waiterEntry) SetPrev(elem *Entry) {
	e.prev = elem
}

// Next returns the entry that follows e in the list.
//
//go:nosplit
func (e *waiterEntry) Next() *Entry {
	return e.next
}

// Prev returns the entry that precedes e in the list.
//
//go:nosplit
func (e *waiterEntry) Prev() *Entry {
	return e.prev
}

// Remove removes e from l.
//
//go:nosplit
func (l *waiterList) Remove(e *Entry) {
	linker := waiterElementMapper{}.linkerFor(e)
	prev := linker.Prev()
	next := linker.Next()

	if prev != nil {
		waiterElementMapper{}.linkerFor(prev).SetNext(next)
	} else if l.head == e {
		l.head = next
	}

	if next != nil {
		waiterElementMapper{}.linkerFor(next).SetPrev(prev)
	} else if l.tail == e {
		l.tail = prev
	}

	linker.SetNext(nil)
	linker.SetPrev(nil)
}

// EventRegister adds a waiter to the wait queue; the waiter will be notified
// when at least one of the events specified in mask happens.
func (q *Queue) EventRegister(e *Entry, mask EventMask) {
	q.mu.Lock()
	e.mask = mask
	q.list.PushBack(e)
	q.mu.Unlock()
}

// EventUnregister removes the given waiter entry from the wait queue.
func (q *Queue) EventUnregister(e *Entry) {
	q.mu.Lock()
	q.list.Remove(e)
	q.mu.Unlock()
}
