package stack

import (
	"github.com/icexin/eggos/gvisor/tcpip"
	"github.com/icexin/eggos/gvisor/waiter"
	"gvisor.dev/gvisor/pkg/buffer"
)

// TransportErrorKind enumerates error types that are handled by the transport
// layer.
type TransportErrorKind int

// TransportEndpoint is the interface that needs to be implemented by transport
// protocol (e.g., tcp, udp) endpoints that can handle packets.
type TransportEndpoint interface {
	// UniqueID returns an unique ID for this transport endpoint.
	UniqueID() uint64

	// HandlePacket is called by the stack when new packets arrive to this
	// transport endpoint. It sets the packet buffer's transport header.
	//
	// HandlePacket takes ownership of the packet.
	HandlePacket(TransportEndpointID, *PacketBuffer)

	// HandleError is called when the transport endpoint receives an error.
	//
	// HandleError takes ownership of the packet buffer.
	HandleError(TransportError, *PacketBuffer)

	// Abort initiates an expedited endpoint teardown. It puts the endpoint
	// in a closed state and frees all resources associated with it. This
	// cleanup may happen asynchronously. Wait can be used to block on this
	// asynchronous cleanup.
	Abort()

	// Wait waits for any worker goroutines owned by the endpoint to stop.
	//
	// An endpoint can be requested to stop its worker goroutines by calling
	// its Close method.
	//
	// Wait will not block if the endpoint hasn't started any goroutines
	// yet, even if it might later.
	Wait()
}

// TransportEndpointID is the identifier of a transport layer protocol endpoint.
//
// +stateify savable
type TransportEndpointID struct {
	// LocalPort is the local port associated with the endpoint.
	LocalPort uint16

	// LocalAddress is the local [network layer] address associated with
	// the endpoint.
	LocalAddress tcpip.Address

	// RemotePort is the remote port associated with the endpoint.
	RemotePort uint16

	// RemoteAddress it the remote [network layer] address associated with
	// the endpoint.
	RemoteAddress tcpip.Address
}

// TransportError is a marker interface for errors that may be handled by the
// transport layer.
type TransportError interface {
	tcpip.SockErrorCause

	// Kind returns the type of the transport error.
	Kind() TransportErrorKind
}

// RawTransportEndpoint is the interface that needs to be implemented by raw
// transport protocol endpoints. RawTransportEndpoints receive the entire
// packet - including the network and transport headers - as delivered to
// netstack.
type RawTransportEndpoint interface {
	// HandlePacket is called by the stack when new packets arrive to
	// this transport endpoint. The packet contains all data from the link
	// layer up.
	//
	// HandlePacket takes ownership of the packet.
	HandlePacket(*PacketBuffer)
}

// TransportProtocol is the interface that needs to be implemented by transport
// protocols (e.g., tcp, udp) that want to be part of the networking stack.
type TransportProtocol interface {
	// Number returns the transport protocol number.
	Number() tcpip.TransportProtocolNumber

	// NewEndpoint creates a new endpoint of the transport protocol.
	NewEndpoint(netProto tcpip.NetworkProtocolNumber, waitQueue *waiter.Queue) (tcpip.Endpoint, tcpip.Error)

	// NewRawEndpoint creates a new raw endpoint of the transport protocol.
	NewRawEndpoint(netProto tcpip.NetworkProtocolNumber, waitQueue *waiter.Queue) (tcpip.Endpoint, tcpip.Error)

	// MinimumPacketSize returns the minimum valid packet size of this
	// transport protocol. The stack automatically drops any packets smaller
	// than this targeted at this protocol.
	MinimumPacketSize() int

	// ParsePorts returns the source and destination ports stored in a
	// packet of this protocol.
	ParsePorts(v buffer.View) (src, dst uint16, err tcpip.Error)

	// HandleUnknownDestinationPacket handles packets targeted at this
	// protocol that don't match any existing endpoint. For example,
	// it is targeted at a port that has no listeners.
	//
	// HandleUnknownDestinationPacket takes ownership of the packet if it handles
	// the issue.
	HandleUnknownDestinationPacket(TransportEndpointID, *PacketBuffer) UnknownDestinationPacketDisposition

	// SetOption allows enabling/disabling protocol specific features.
	// SetOption returns an error if the option is not supported or the
	// provided option value is invalid.
	SetOption(option tcpip.SettableTransportProtocolOption) tcpip.Error

	// Option allows retrieving protocol specific option values.
	// Option returns an error if the option is not supported or the
	// provided option value is invalid.
	Option(option tcpip.GettableTransportProtocolOption) tcpip.Error

	// Close requests that any worker goroutines owned by the protocol
	// stop.
	Close()

	// Wait waits for any worker goroutines owned by the protocol to stop.
	Wait()

	// Parse sets pkt.TransportHeader and trims pkt.Data appropriately. It does
	// neither and returns false if pkt.Data is too small, i.e. pkt.Data.Size() <
	// MinimumPacketSize()
	Parse(pkt *PacketBuffer) (ok bool)
}

// UnknownDestinationPacketDisposition enumerates the possible return values from
// HandleUnknownDestinationPacket().
type UnknownDestinationPacketDisposition int

// NetworkInterface is a network interface.
type NetworkInterface interface {
	NetworkLinkEndpoint

	// ID returns the interface's ID.
	ID() tcpip.NICID

	// IsLoopback returns true if the interface is a loopback interface.
	IsLoopback() bool

	// Name returns the name of the interface.
	//
	// May return an empty string if the interface is not configured with a name.
	Name() string

	// Enabled returns true if the interface is enabled.
	Enabled() bool

	// Promiscuous returns true if the interface is in promiscuous mode.
	//
	// When in promiscuous mode, the interface should accept all packets.
	Promiscuous() bool

	// Spoofing returns true if the interface is in spoofing mode.
	//
	// When in spoofing mode, the interface should consider all addresses as
	// assigned to it.
	Spoofing() bool

	// PrimaryAddress returns the primary address associated with the interface.
	//
	// PrimaryAddress will return the first non-deprecated address if such an
	// address exists. If no non-deprecated addresses exist, the first deprecated
	// address will be returned. If no deprecated addresses exist, the zero value
	// will be returned.
	PrimaryAddress(tcpip.NetworkProtocolNumber) (tcpip.AddressWithPrefix, tcpip.Error)

	// CheckLocalAddress returns true if the address exists on the interface.
	CheckLocalAddress(tcpip.NetworkProtocolNumber, tcpip.Address) bool

	// WritePacketToRemote writes the packet to the given remote link address.
	WritePacketToRemote(tcpip.LinkAddress, tcpip.NetworkProtocolNumber, *PacketBuffer) tcpip.Error

	// WritePacket writes a packet with the given protocol through the given
	// route.
	//
	// WritePacket takes ownership of the packet buffer. The packet buffer's
	// network and transport header must be set.
	WritePacket(*Route, tcpip.NetworkProtocolNumber, *PacketBuffer) tcpip.Error

	// WritePackets writes packets with the given protocol through the given
	// route. Must not be called with an empty list of packet buffers.
	//
	// WritePackets takes ownership of the packet buffers.
	//
	// Right now, WritePackets is used only when the software segmentation
	// offload is enabled. If it will be used for something else, syscall filters
	// may need to be updated.
	WritePackets(*Route, PacketBufferList, tcpip.NetworkProtocolNumber) (int, tcpip.Error)

	// HandleNeighborProbe processes an incoming neighbor probe (e.g. ARP
	// request or NDP Neighbor Solicitation).
	//
	// HandleNeighborProbe assumes that the probe is valid for the network
	// interface the probe was received on.
	HandleNeighborProbe(tcpip.NetworkProtocolNumber, tcpip.Address, tcpip.LinkAddress) tcpip.Error

	// HandleNeighborConfirmation processes an incoming neighbor confirmation
	// (e.g. ARP reply or NDP Neighbor Advertisement).
	HandleNeighborConfirmation(tcpip.NetworkProtocolNumber, tcpip.Address, tcpip.LinkAddress, ReachabilityConfirmationFlags) tcpip.Error
}

// NetworkProtocol is the interface that needs to be implemented by network
// protocols (e.g., ipv4, ipv6) that want to be part of the networking stack.
type NetworkProtocol interface {
	// Number returns the network protocol number.
	Number() tcpip.NetworkProtocolNumber

	// MinimumPacketSize returns the minimum valid packet size of this
	// network protocol. The stack automatically drops any packets smaller
	// than this targeted at this protocol.
	MinimumPacketSize() int

	// DefaultPrefixLen returns the protocol's default prefix length.
	DefaultPrefixLen() int

	// ParseAddresses returns the source and destination addresses stored in a
	// packet of this protocol.
	ParseAddresses(v buffer.View) (src, dst tcpip.Address)

	// NewEndpoint creates a new endpoint of this protocol.
	NewEndpoint(nic NetworkInterface, dispatcher TransportDispatcher) NetworkEndpoint

	// SetOption allows enabling/disabling protocol specific features.
	// SetOption returns an error if the option is not supported or the
	// provided option value is invalid.
	SetOption(option tcpip.SettableNetworkProtocolOption) tcpip.Error

	// Option allows retrieving protocol specific option values.
	// Option returns an error if the option is not supported or the
	// provided option value is invalid.
	Option(option tcpip.GettableNetworkProtocolOption) tcpip.Error

	// Close requests that any worker goroutines owned by the protocol
	// stop.
	Close()

	// Wait waits for any worker goroutines owned by the protocol to stop.
	Wait()

	// Parse sets pkt.NetworkHeader and trims pkt.Data appropriately. It
	// returns:
	// - The encapsulated protocol, if present.
	// - Whether there is an encapsulated transport protocol payload (e.g. ARP
	//   does not encapsulate anything).
	// - Whether pkt.Data was large enough to parse and set pkt.NetworkHeader.
	Parse(pkt *PacketBuffer) (proto tcpip.TransportProtocolNumber, hasTransportHdr bool, ok bool)
}
