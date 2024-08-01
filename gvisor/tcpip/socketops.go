package tcpip

import (
	"sync"
	"sync/atomic"
)

// SockError represents a queue entry in the per-socket error queue.
//
// +stateify savable
type SockError struct {
	sockErrorEntry

	// Err is the error caused by the errant packet.
	Err Error
	// Cause is the detailed cause of the error.
	Cause SockErrorCause

	// Payload is the errant packet's payload.
	Payload []byte
	// Dst is the original destination address of the errant packet.
	Dst FullAddress
	// Offender is the original sender address of the errant packet.
	Offender FullAddress
	// NetProto is the network protocol being used to transmit the packet.
	NetProto NetworkProtocolNumber
}

// Entry is a default implementation of Linker. Users can add anonymous fields
// of this type to their structs to make them automatically implement the
// methods needed by List.
//
// +stateify savable
type sockErrorEntry struct {
	next *SockError
	prev *SockError
}

// SocketOptions contains all the variables which store values for SOL_SOCKET,
// SOL_IP, SOL_IPV6 and SOL_TCP level options.
//
// +stateify savable
type SocketOptions struct {
	handler SocketOptionsHandler

	// StackHandler is initialized at the creation time and will not change.
	stackHandler StackHandler `state:"manual"`

	// These fields are accessed and modified using atomic operations.

	// broadcastEnabled determines whether datagram sockets are allowed to
	// send packets to a broadcast address.
	broadcastEnabled uint32

	// passCredEnabled determines whether SCM_CREDENTIALS socket control
	// messages are enabled.
	passCredEnabled uint32

	// noChecksumEnabled determines whether UDP checksum is disabled while
	// transmitting for this socket.
	noChecksumEnabled uint32

	// reuseAddressEnabled determines whether Bind() should allow reuse of
	// local address.
	reuseAddressEnabled uint32

	// reusePortEnabled determines whether to permit multiple sockets to be
	// bound to an identical socket address.
	reusePortEnabled uint32

	// keepAliveEnabled determines whether TCP keepalive is enabled for this
	// socket.
	keepAliveEnabled uint32

	// multicastLoopEnabled determines whether multicast packets sent over a
	// non-loopback interface will be looped back.
	multicastLoopEnabled uint32

	// receiveTOSEnabled is used to specify if the TOS ancillary message is
	// passed with incoming packets.
	receiveTOSEnabled uint32

	// receiveTClassEnabled is used to specify if the IPV6_TCLASS ancillary
	// message is passed with incoming packets.
	receiveTClassEnabled uint32

	// receivePacketInfoEnabled is used to specify if more inforamtion is
	// provided with incoming packets such as interface index and address.
	receivePacketInfoEnabled uint32

	// hdrIncludeEnabled is used to indicate for a raw endpoint that all packets
	// being written have an IP header and the endpoint should not attach an IP
	// header.
	hdrIncludedEnabled uint32

	// v6OnlyEnabled is used to determine whether an IPv6 socket is to be
	// restricted to sending and receiving IPv6 packets only.
	v6OnlyEnabled uint32

	// quickAckEnabled is used to represent the value of TCP_QUICKACK option.
	// It currently does not have any effect on the TCP endpoint.
	quickAckEnabled uint32

	// delayOptionEnabled is used to specify if data should be sent out immediately
	// by the transport protocol. For TCP, it determines if the Nagle algorithm
	// is on or off.
	delayOptionEnabled uint32

	// corkOptionEnabled is used to specify if data should be held until segments
	// are full by the TCP transport protocol.
	corkOptionEnabled uint32

	// receiveOriginalDstAddress is used to specify if the original destination of
	// the incoming packet should be returned as an ancillary message.
	receiveOriginalDstAddress uint32

	// recvErrEnabled determines whether extended reliable error message passing
	// is enabled.
	recvErrEnabled uint32

	// errQueue is the per-socket error queue. It is protected by errQueueMu.
	errQueueMu sync.Mutex `state:"nosave"`
	errQueue   sockErrorList

	// bindToDevice determines the device to which the socket is bound.
	bindToDevice int32

	// getSendBufferLimits provides the handler to get the min, default and
	// max size for send buffer. It  is initialized at the creation time and
	// will not change.
	getSendBufferLimits GetSendBufferLimits `state:"manual"`

	// sendBufSizeMu protects sendBufferSize and calls to
	// handler.OnSetSendBufferSize.
	sendBufSizeMu sync.Mutex `state:"nosave"`

	// sendBufferSize determines the send buffer size for this socket.
	sendBufferSize int64

	// getReceiveBufferLimits provides the handler to get the min, default and
	// max size for receive buffer. It is initialized at the creation time and
	// will not change.
	getReceiveBufferLimits GetReceiveBufferLimits `state:"manual"`

	// receiveBufSizeMu protects receiveBufferSize and calls to
	// handler.OnSetReceiveBufferSize.
	receiveBufSizeMu sync.Mutex `state:"nosave"`

	// receiveBufferSize determines the receive buffer size for this socket.
	receiveBufferSize int64

	// mu protects the access to the below fields.
	mu sync.Mutex `state:"nosave"`

	// linger determines the amount of time the socket should linger before
	// close. We currently implement this option for TCP socket only.
	linger LingerOption
}

// SendBufferSizeOption is used by stack.(Stack*).Option/SetOption to
// get/set the default, min and max send buffer sizes.
type SendBufferSizeOption struct {
	// Min is the minimum size for send buffer.
	Min int

	// Default is the default size for send buffer.
	Default int

	// Max is the maximum size for send buffer.
	Max int
}

// ReceiveBufferSizeOption is used by stack.(Stack*).Option/SetOption to
// get/set the default, min and max receive buffer sizes.
type ReceiveBufferSizeOption struct {
	// Min is the minimum size for send buffer.
	Min int

	// Default is the default size for send buffer.
	Default int

	// Max is the maximum size for send buffer.
	Max int
}

// GetSendBufferLimits is used to get the send buffer size limits.
type GetSendBufferLimits func(StackHandler) SendBufferSizeOption

// GetReceiveBufferLimits is used to get the send buffer size limits.
type GetReceiveBufferLimits func(StackHandler) ReceiveBufferSizeOption

func storeAtomicBool(addr *uint32, v bool) {
	var val uint32
	if v {
		val = 1
	}
	atomic.StoreUint32(addr, val)
}

// StackHandler holds methods to access the stack options. These must be
// implemented by the stack.
type StackHandler interface {
	// Option allows retrieving stack wide options.
	Option(option interface{}) Error

	// TransportProtocolOption allows retrieving individual protocol level
	// option values.
	TransportProtocolOption(proto TransportProtocolNumber, option GettableTransportProtocolOption) Error
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
type sockErrorList struct {
	head *SockError
	tail *SockError
}

// SockErrOrigin represents the constants for error origin.
type SockErrOrigin uint8

// SockErrorCause is the cause of a socket error.
type SockErrorCause interface {
	// Origin is the source of the error.
	Origin() SockErrOrigin

	// Type is the origin specific type of error.
	Type() uint8

	// Code is the origin and type specific error code.
	Code() uint8

	// Info is any extra information about the error.
	Info() uint32
}

// SetKeepAlive sets value for SO_KEEPALIVE option.
func (so *SocketOptions) SetKeepAlive(v bool) {
	storeAtomicBool(&so.keepAliveEnabled, v)
	so.handler.OnKeepAliveSet(v)
}

// SetReuseAddress sets value for SO_REUSEADDR option.
func (so *SocketOptions) SetReuseAddress(v bool) {
	storeAtomicBool(&so.reuseAddressEnabled, v)
	so.handler.OnReuseAddressSet(v)
}

// SetBroadcast sets value for SO_BROADCAST option.
func (so *SocketOptions) SetBroadcast(v bool) {
	storeAtomicBool(&so.broadcastEnabled, v)
}

// SetDelayOption sets inverted value for TCP_NODELAY option.
func (so *SocketOptions) SetDelayOption(v bool) {
	storeAtomicBool(&so.delayOptionEnabled, v)
	so.handler.OnDelayOptionSet(v)
}

// GetLastError gets value for SO_ERROR option.
func (so *SocketOptions) GetLastError() Error {
	return so.handler.LastError()
}

// SockOptInt represents socket options which values have the int type.
type SockOptInt int

// SocketOptionsHandler holds methods that help define endpoint specific
// behavior for socket level socket options. These must be implemented by
// endpoints to get notified when socket level options are set.
type SocketOptionsHandler interface {
	// OnReuseAddressSet is invoked when SO_REUSEADDR is set for an endpoint.
	OnReuseAddressSet(v bool)

	// OnReusePortSet is invoked when SO_REUSEPORT is set for an endpoint.
	OnReusePortSet(v bool)

	// OnKeepAliveSet is invoked when SO_KEEPALIVE is set for an endpoint.
	OnKeepAliveSet(v bool)

	// OnDelayOptionSet is invoked when TCP_NODELAY is set for an endpoint.
	// Note that v will be the inverse of TCP_NODELAY option.
	OnDelayOptionSet(v bool)

	// OnCorkOptionSet is invoked when TCP_CORK is set for an endpoint.
	OnCorkOptionSet(v bool)

	// LastError is invoked when SO_ERROR is read for an endpoint.
	LastError() Error

	// UpdateLastError updates the endpoint specific last error field.
	UpdateLastError(err Error)

	// HasNIC is invoked to check if the NIC is valid for SO_BINDTODEVICE.
	HasNIC(v int32) bool

	// OnSetSendBufferSize is invoked when the send buffer size for an endpoint is
	// changed. The handler is invoked with the new value for the socket send
	// buffer size. It also returns the newly set value.
	OnSetSendBufferSize(v int64) (newSz int64)

	// OnSetReceiveBufferSize is invoked by SO_RCVBUF and SO_RCVBUFFORCE.
	OnSetReceiveBufferSize(v, oldSz int64) (newSz int64)
}
