package stack

// TableID identifies a specific table.
type TableID int

// Each value identifies a specific table.
const (
	NATID TableID = iota
	MangleID
	FilterID
	NumTables
)
