package tcpip

import "fmt"

// ErrNoRoute indicates the operation is not able to find a route to the
// destination.
//
// +stateify savable
type ErrNoRoute struct{}

func (*ErrNoRoute) isError() {}

// IgnoreStats implements Error.
func (*ErrNoRoute) IgnoreStats() bool {
	return false
}
func (*ErrNoRoute) String() string { return "no route" }

// ErrConnectionRefused indicates the connection was refused.
//
// +stateify savable
type ErrConnectionRefused struct{}

func (*ErrConnectionRefused) isError() {}

// IgnoreStats implements Error.
func (*ErrConnectionRefused) IgnoreStats() bool {
	return false
}
func (*ErrConnectionRefused) String() string { return "connection was refused" }

// ErrConnectStarted indicates the endpoint is connecting asynchronously.
//
// +stateify savable
type ErrConnectStarted struct{}

func (*ErrConnectStarted) isError() {}

// IgnoreStats implements Error.
func (*ErrConnectStarted) IgnoreStats() bool {
	return true
}
func (*ErrConnectStarted) String() string { return "connection attempt started" }

// ErrClosedForSend indicates the endpoint is closed for outgoing data.
//
// +stateify savable
type ErrClosedForSend struct{}

func (*ErrClosedForSend) isError() {}

// IgnoreStats implements Error.
func (*ErrClosedForSend) IgnoreStats() bool {
	return false
}
func (*ErrClosedForSend) String() string { return "endpoint is closed for send" }

// Error represents an error in the netstack error space.
//
// The error interface is intentionally omitted to avoid loss of type
// information that would occur if these errors were passed as error.
type Error interface {
	isError()

	// IgnoreStats indicates whether this error should be included in failure
	// counts in tcpip.Stats structs.
	IgnoreStats() bool

	fmt.Stringer
}

// ErrWouldBlock indicates the operation would block.
//
// +stateify savable
type ErrWouldBlock struct{}

func (*ErrWouldBlock) isError() {}

// IgnoreStats implements Error.
func (*ErrWouldBlock) IgnoreStats() bool {
	return true
}
func (*ErrWouldBlock) String() string { return "operation would block" }

// ErrClosedForReceive indicates the endpoint is closed for incoming data.
//
// +stateify savable
type ErrClosedForReceive struct{}

func (*ErrClosedForReceive) isError() {}

// IgnoreStats implements Error.
func (*ErrClosedForReceive) IgnoreStats() bool {
	return false
}
func (*ErrClosedForReceive) String() string { return "endpoint is closed for receive" }

// ErrDuplicateAddress indicates the operation encountered a duplicate address.
//
// +stateify savable
type ErrDuplicateAddress struct{}

func (*ErrDuplicateAddress) isError() {}

// IgnoreStats implements Error.
func (*ErrDuplicateAddress) IgnoreStats() bool {
	return false
}
func (*ErrDuplicateAddress) String() string { return "duplicate address" }

// ErrUnknownNICID indicates an unknown NIC ID was provided.
//
// +stateify savable
type ErrUnknownNICID struct{}

func (*ErrUnknownNICID) isError() {}

// IgnoreStats implements Error.
func (*ErrUnknownNICID) IgnoreStats() bool {
	return false
}
func (*ErrUnknownNICID) String() string { return "unknown nic id" }

// ErrUnknownProtocol indicates an unknown protocol was requested.
//
// +stateify savable
type ErrUnknownProtocol struct{}

func (*ErrUnknownProtocol) isError() {}

// IgnoreStats implements Error.
func (*ErrUnknownProtocol) IgnoreStats() bool {
	return false
}
func (*ErrUnknownProtocol) String() string { return "unknown protocol" }
