package buffer

const (
	// embeddedCount is the number of buffer structures embedded in the pool. It
	// is also the number for overflow allocations.
	embeddedCount = 8

	// defaultBufferSize is the default size for each underlying storage buffer.
	//
	// It is slightly less than two pages. This is done intentionally to ensure
	// that the buffer object aligns with runtime internals. This two page size
	// will effectively minimize internal fragmentation, but still have a large
	// enough chunk to limit excessive segmentation.
	defaultBufferSize = 8144
)

// pool allocates buffer.
//
// It contains an embedded buffer storage for fast path when the number of
// buffers needed is small.
//
// +stateify savable
type pool struct {
	bufferSize      int
	avail           []buffer              `state:"nosave"`
	embeddedStorage [embeddedCount]buffer `state:"wait"`
}
