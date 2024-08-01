package stack

import "sync"

// ConnTrack tracks all connections created for NAT rules. Most users are
// expected to only call handlePacket, insertRedirectConn, and maybeInsertNoop.
//
// ConnTrack keeps all connections in a slice of buckets, each of which holds a
// linked list of tuples. This gives us some desirable properties:
//   - Each bucket has its own lock, lessening lock contention.
//   - The slice is large enough that lists stay short (<10 elements on average).
//     Thus traversal is fast.
//   - During linked list traversal we reap expired connections. This amortizes
//     the cost of reaping them and makes reapUnused faster.
//
// Locks are ordered by their location in the buckets slice. That is, a
// goroutine that locks buckets[i] can only lock buckets[j] s.t. i < j.
//
// +stateify savable
type ConnTrack struct {
	// seed is a one-time random value initialized at stack startup
	// and is used in the calculation of hash keys for the list of buckets.
	// It is immutable.
	seed uint32

	// mu protects the buckets slice, but not buckets' contents. Only take
	// the write lock if you are modifying the slice or saving for S/R.
	mu sync.RWMutex `state:"nosave"`

	// buckets is protected by mu.
	buckets []bucket
}

// +stateify savable
type bucket struct {
	// mu protects tuples.
	mu     sync.Mutex `state:"nosave"`
	tuples tupleList
}
