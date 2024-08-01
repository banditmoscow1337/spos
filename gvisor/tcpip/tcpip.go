package tcpip

import (
	"bytes"
	"io"
	"time"

	"github.com/icexin/eggos/gvisor/waiter"
)

var _ Payloader = (*bytes.Buffer)(nil)
var _ Payloader = (*bytes.Reader)(nil)

var _ io.Writer = (*SliceWriter)(nil)

type InetAddr [4]byte

// SliceWriter implements io.Writer for slices.
type SliceWriter []byte

// Write implements io.Writer.Write.
func (s *SliceWriter) Write(b []byte) (int, error) {
	n := copy(*s, b)
	*s = (*s)[n:]
	var err error
	if n != len(b) {
		err = io.ErrShortWrite
	}
	return n, err
}

type SockAddrInet struct {
	Family uint16
	Port   uint16
	Addr   InetAddr
	_      [8]uint8 // pad to sizeof(struct sockaddr).
}

// LingerOption is used by SetSockOpt/GetSockOpt to set/get the
// duration for which a socket lingers before returning from Close.
//
// +marshal
// +stateify savable
type LingerOption struct {
	Enabled bool
	Timeout time.Duration
}

// NICID is a number that uniquely identifies a NIC.
type NICID int32

// Address is a byte slice cast as a string that represents the address of a
// network node. Or, in the case of unix endpoints, it may represent a path.
type Address string

// FullAddress represents a full transport node address, as required by the
// Connect() and Bind() methods.
//
// +stateify savable
type FullAddress struct {
	// NIC is the ID of the NIC this address refers to.
	//
	// This may not be used by all endpoint types.
	NIC NICID

	// Addr is the network or link layer address.
	Addr Address

	// Port is the transport port.
	//
	// This may not be used by all endpoint types.
	Port uint16
}

// IPPacketInfo is the message structure for IP_PKTINFO.
//
// +stateify savable
type IPPacketInfo struct {
	// NIC is the ID of the NIC to be used.
	NIC NICID

	// LocalAddr is the local address.
	LocalAddr Address

	// DestinationAddr is the destination address found in the IP header.
	DestinationAddr Address
}

// Payloader is an interface that provides data.
//
// This interface allows the endpoint to request the amount of data it needs
// based on internal buffers without exposing them.
type Payloader interface {
	io.Reader

	// Len returns the number of bytes of the unread portion of the
	// Reader.
	Len() int
}

// Endpoint is the interface implemented by transport protocols (e.g., tcp, udp)
// that exposes functionality like read, write, connect, etc. to users of the
// networking stack.
type Endpoint interface {
	// Close puts the endpoint in a closed state and frees all resources
	// associated with it. Close initiates the teardown process, the
	// Endpoint may not be fully closed when Close returns.
	Close()

	// Abort initiates an expedited endpoint teardown. As compared to
	// Close, Abort prioritizes closing the Endpoint quickly over cleanly.
	// Abort is best effort; implementing Abort with Close is acceptable.
	Abort()

	// Read reads data from the endpoint and optionally writes to dst.
	//
	// This method does not block if there is no data pending; in this case,
	// ErrWouldBlock is returned.
	//
	// If non-zero number of bytes are successfully read and written to dst, err
	// must be nil. Otherwise, if dst failed to write anything, ErrBadBuffer
	// should be returned.
	Read(io.Writer, ReadOptions) (ReadResult, Error)

	// Write writes data to the endpoint's peer. This method does not block if
	// the data cannot be written.
	//
	// Unlike io.Writer.Write, Endpoint.Write transfers ownership of any bytes
	// successfully written to the Endpoint. That is, if a call to
	// Write(SlicePayload{data}) returns (n, err), it may retain data[:n], and
	// the caller should not use data[:n] after Write returns.
	//
	// Note that unlike io.Writer.Write, it is not an error for Write to
	// perform a partial write (if n > 0, no error may be returned). Only
	// stream (TCP) Endpoints may return partial writes, and even then only
	// in the case where writing additional data would block. Other Endpoints
	// will either write the entire message or return an error.
	Write(Payloader, WriteOptions) (int64, Error)

	// Connect connects the endpoint to its peer. Specifying a NIC is
	// optional.
	//
	// There are three classes of return values:
	//	nil -- the attempt to connect succeeded.
	//	ErrConnectStarted/ErrAlreadyConnecting -- the connect attempt started
	//		but hasn't completed yet. In this case, the caller must call Connect
	//		or GetSockOpt(ErrorOption) when the endpoint becomes writable to
	//		get the actual result. The first call to Connect after the socket has
	//		connected returns nil. Calling connect again results in ErrAlreadyConnected.
	//	Anything else -- the attempt to connect failed.
	//
	// If address.Addr is empty, this means that Endpoint has to be
	// disconnected if this is supported, otherwise
	// ErrAddressFamilyNotSupported must be returned.
	Connect(address FullAddress) Error

	// Disconnect disconnects the endpoint from its peer.
	Disconnect() Error

	// Shutdown closes the read and/or write end of the endpoint connection
	// to its peer.
	Shutdown(flags ShutdownFlags) Error

	// Listen puts the endpoint in "listen" mode, which allows it to accept
	// new connections.
	Listen(backlog int) Error

	// Accept returns a new endpoint if a peer has established a connection
	// to an endpoint previously set to listen mode. This method does not
	// block if no new connections are available.
	//
	// The returned Queue is the wait queue for the newly created endpoint.
	//
	// If peerAddr is not nil then it is populated with the peer address of the
	// returned endpoint.
	Accept(peerAddr *FullAddress) (Endpoint, *waiter.Queue, Error)

	// Bind binds the endpoint to a specific local address and port.
	// Specifying a NIC is optional.
	Bind(address FullAddress) Error

	// GetLocalAddress returns the address to which the endpoint is bound.
	GetLocalAddress() (FullAddress, Error)

	// GetRemoteAddress returns the address to which the endpoint is
	// connected.
	GetRemoteAddress() (FullAddress, Error)

	// Readiness returns the current readiness of the endpoint. For example,
	// if waiter.EventIn is set, the endpoint is immediately readable.
	Readiness(mask waiter.EventMask) waiter.EventMask

	// SetSockOpt sets a socket option.
	SetSockOpt(opt SettableSocketOption) Error

	// SetSockOptInt sets a socket option, for simple cases where a value
	// has the int type.
	SetSockOptInt(opt SockOptInt, v int) Error

	// GetSockOpt gets a socket option.
	GetSockOpt(opt GettableSocketOption) Error

	// GetSockOptInt gets a socket option for simple cases where a return
	// value has the int type.
	GetSockOptInt(SockOptInt) (int, Error)

	// State returns a socket's lifecycle state. The returned value is
	// protocol-specific and is primarily used for diagnostics.
	State() uint32

	// ModerateRecvBuf should be called everytime data is copied to the user
	// space. This allows for dynamic tuning of recv buffer space for a
	// given socket.
	//
	// NOTE: This method is a no-op for sockets other than TCP.
	ModerateRecvBuf(copied int)

	// Info returns a copy to the transport endpoint info.
	Info() EndpointInfo

	// Stats returns a reference to the endpoint stats.
	Stats() EndpointStats

	// SetOwner sets the task owner to the endpoint owner.
	SetOwner(owner PacketOwner)

	// LastError clears and returns the last error reported by the endpoint.
	LastError() Error

	// SocketOptions returns the structure which contains all the socket
	// level options.
	SocketOptions() *SocketOptions
}

// AddressMask is a bitmask for an address.
type AddressMask string

// EndpointInfo is the interface implemented by each endpoint info struct.
type EndpointInfo interface {
	// IsEndpointInfo is an empty method to implement the tcpip.EndpointInfo
	// marker interface.
	IsEndpointInfo()
}

// WriteOptions contains options for Endpoint.Write.
type WriteOptions struct {
	// If To is not nil, write to the given address instead of the endpoint's
	// peer.
	To *FullAddress

	// More has the same semantics as Linux's MSG_MORE.
	More bool

	// EndOfRecord has the same semantics as Linux's MSG_EOR.
	EndOfRecord bool

	// Atomic means that all data fetched from Payloader must be written to the
	// endpoint. If Atomic is false, then data fetched from the Payloader may be
	// discarded if available endpoint buffer space is unsufficient.
	Atomic bool
}

// A ControlMessages contains socket control messages for IP sockets.
//
// +stateify savable
type ControlMessages struct {
	// HasTimestamp indicates whether Timestamp is valid/set.
	HasTimestamp bool

	// Timestamp is the time (in ns) that the last packet used to create
	// the read data was received.
	Timestamp int64

	// HasInq indicates whether Inq is valid/set.
	HasInq bool

	// Inq is the number of bytes ready to be received.
	Inq int32

	// HasTOS indicates whether Tos is valid/set.
	HasTOS bool

	// TOS is the IPv4 type of service of the associated packet.
	TOS uint8

	// HasTClass indicates whether TClass is valid/set.
	HasTClass bool

	// TClass is the IPv6 traffic class of the associated packet.
	TClass uint32

	// HasIPPacketInfo indicates whether PacketInfo is set.
	HasIPPacketInfo bool

	// PacketInfo holds interface and address data on an incoming packet.
	PacketInfo IPPacketInfo

	// HasOriginalDestinationAddress indicates whether OriginalDstAddress is
	// set.
	HasOriginalDstAddress bool

	// OriginalDestinationAddress holds the original destination address
	// and port of the incoming packet.
	OriginalDstAddress FullAddress

	// SockErr is the dequeued socket error on recvmsg(MSG_ERRQUEUE).
	SockErr *SockError
}

// ReadResult represents result for a successful Endpoint.Read.
type ReadResult struct {
	// Count is the number of bytes received and written to the buffer.
	Count int

	// Total is the number of bytes of the received packet. This can be used to
	// determine whether the read is truncated.
	Total int

	// ControlMessages is the control messages received.
	ControlMessages ControlMessages

	// RemoteAddr is the remote address if ReadOptions.NeedAddr is true.
	RemoteAddr FullAddress

	// LinkPacketInfo is the link-layer information of the received packet if
	// ReadOptions.NeedLinkPacketInfo is true.
	LinkPacketInfo LinkPacketInfo
}

// KeepaliveIdleOption is used by SetSockOpt/GetSockOpt to specify the time a
// connection must remain idle before the first TCP keepalive packet is sent.
// Once this time is reached, KeepaliveIntervalOption is used instead.
type KeepaliveIdleOption time.Duration

func (*KeepaliveIdleOption) isGettableSocketOption() {}

func (*KeepaliveIdleOption) isSettableSocketOption() {}

// KeepaliveIntervalOption is used by SetSockOpt/GetSockOpt to specify the
// interval between sending TCP keepalive packets.
type KeepaliveIntervalOption time.Duration

func (*KeepaliveIntervalOption) isGettableSocketOption() {}

func (*KeepaliveIntervalOption) isSettableSocketOption() {}

// LinkPacketInfo holds Link layer information for a received packet.
//
// +stateify savable
type LinkPacketInfo struct {
	// Protocol is the NetworkProtocolNumber for the packet.
	Protocol NetworkProtocolNumber

	// PktType is used to indicate the destination of the packet.
	PktType PacketType
}

// NetworkProtocolNumber is the EtherType of a network protocol in an Ethernet
// frame.
//
// See: https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml
type NetworkProtocolNumber uint32

// PacketType is used to indicate the destination of the packet.
type PacketType uint8

// PacketOwner is used to get UID and GID of the packet.
type PacketOwner interface {
	// UID returns KUID of the packet.
	KUID() uint32

	// GID returns KGID of the packet.
	KGID() uint32
}

// EndpointStats is the interface implemented by each endpoint stats struct.
type EndpointStats interface {
	// IsEndpointStats is an empty method to implement the tcpip.EndpointStats
	// marker interface.
	IsEndpointStats()
}

// GettableTransportProtocolOption is a marker interface for transport protocol
// options that may be queried.
type GettableTransportProtocolOption interface {
	isGettableTransportProtocolOption()
}

// ReadOptions contains options for Endpoint.Read.
type ReadOptions struct {
	// Peek indicates whether this read is a peek.
	Peek bool

	// NeedRemoteAddr indicates whether to return the remote address, if
	// supported.
	NeedRemoteAddr bool

	// NeedLinkPacketInfo indicates whether to return the link-layer information,
	// if supported.
	NeedLinkPacketInfo bool
}

// ShutdownFlags represents flags that can be passed to the Shutdown() method
// of the Endpoint interface.
type ShutdownFlags int

// SettableSocketOption is a marker interface for socket options that may be
// configured.
type SettableSocketOption interface {
	isSettableSocketOption()
}

// GettableSocketOption is a marker interface for socket options that may be
// queried.
type GettableSocketOption interface {
	isGettableSocketOption()
}

// TransportProtocolNumber is the number of a transport protocol.
type TransportProtocolNumber uint32

// LinkAddress is a byte slice cast as a string that represents a link address.
// It is typically a 6-byte MAC address.
type LinkAddress string

// Route is a row in the routing table. It specifies through which NIC (and
// gateway) sets of packets should be routed. A row is considered viable if the
// masked target address matches the destination address in the row.
type Route struct {
	// Destination must contain the target address for this row to be viable.
	Destination Subnet

	// Gateway is the gateway to be used if this row is viable.
	Gateway Address

	// NIC is the id of the nic to be used if this row is viable.
	NIC NICID
}

// Subnet is a subnet defined by its address and mask.
type Subnet struct {
	address Address
	mask    AddressMask
}

// SettableTransportProtocolOption is a marker interface for transport protocol
// options that may be set.
type SettableTransportProtocolOption interface {
	isSettableTransportProtocolOption()
}

// AddressWithPrefix is an address with its subnet prefix length.
type AddressWithPrefix struct {
	// Address is a network address.
	Address Address

	// PrefixLen is the subnet prefix length.
	PrefixLen int
}

// ProtocolAddress is an address and the network protocol it is associated
// with.
type ProtocolAddress struct {
	// Protocol is the protocol of the address.
	Protocol NetworkProtocolNumber

	// AddressWithPrefix is a network address with its subnet prefix length.
	AddressWithPrefix AddressWithPrefix
}

// A Clock provides the current time and schedules work for execution.
//
// Times returned by a Clock should always be used for application-visible
// time. Only monotonic times should be used for netstack internal timekeeping.
type Clock interface {
	// Now returns the current local time.
	Now() time.Time

	// NowMonotonic returns the current monotonic clock reading.
	NowMonotonic() MonotonicTime

	// AfterFunc waits for the duration to elapse and then calls f in its own
	// goroutine. It returns a Timer that can be used to cancel the call using
	// its Stop method.
	AfterFunc(d time.Duration, f func()) Timer
}

// MonotonicTime is a monotonic clock reading.
//
// +stateify savable
type MonotonicTime struct {
	nanoseconds int64
}

// Timer represents a single event. A Timer must be created with
// Clock.AfterFunc.
type Timer interface {
	// Stop prevents the Timer from firing. It returns true if the call stops the
	// timer, false if the timer has already expired or been stopped.
	//
	// If Stop returns false, then the timer has already expired and the function
	// f of Clock.AfterFunc(d, f) has been started in its own goroutine; Stop
	// does not wait for f to complete before returning. If the caller needs to
	// know whether f is completed, it must coordinate with f explicitly.
	Stop() bool

	// Reset changes the timer to expire after duration d.
	//
	// Reset should be invoked only on stopped or expired timers. If the timer is
	// known to have expired, Reset can be used directly. Otherwise, the caller
	// must coordinate with the function f of Clock.AfterFunc(d, f).
	Reset(d time.Duration)
}
