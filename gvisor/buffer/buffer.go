package buffer

// buffer encapsulates a queueable byte buffer.
//
// +stateify savable
type buffer struct {
	data  []byte
	read  int
	write int
	bufferEntry
}
