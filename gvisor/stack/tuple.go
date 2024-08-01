package stack

// tuple holds a connection's identifying and manipulating data in one
// direction. It is immutable.
//
// +stateify savable
type tuple struct {
	// tupleEntry is used to build an intrusive list of tuples.
	tupleEntry

	tupleID

	// conn is the connection tracking entry this tuple belongs to.
	conn *conn

	// direction is the direction of the tuple.
	direction direction
}
