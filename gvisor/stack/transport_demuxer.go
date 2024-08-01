package stack

import (
	"sync"

	"github.com/icexin/eggos/gvisor/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/ports"
)

// queuedTransportProtocol if supported by a protocol implementation will cause
// the dispatcher to delivery packets to the QueuePacket method instead of
// calling HandlePacket directly on the endpoint.
type queuedTransportProtocol interface {
	QueuePacket(ep TransportEndpoint, id TransportEndpointID, pkt *PacketBuffer)
}

// transportDemuxer demultiplexes packets targeted at a transport endpoint
// (i.e., after they've been parsed by the network layer). It does two levels
// of demultiplexing: first based on the network and transport protocols, then
// based on endpoints IDs. It should only be instantiated via
// newTransportDemuxer.
type transportDemuxer struct {
	stack *Stack

	// protocol is immutable.
	protocol        map[protocolIDs]*transportEndpoints
	queuedProtocols map[protocolIDs]queuedTransportProtocol
}

type protocolIDs struct {
	network   tcpip.NetworkProtocolNumber
	transport tcpip.TransportProtocolNumber
}

// multiPortEndpoint is a container for TransportEndpoints which are bound to
// the same pair of address and port. endpointsArr always has at least one
// element.
//
// FIXME(gvisor.dev/issue/873): Restore this properly. Currently, we just save
// this to ensure that the underlying endpoints get saved/restored, but not not
// use the restored copy.
//
// +stateify savable
type multiPortEndpoint struct {
	mu         sync.RWMutex `state:"nosave"`
	demux      *transportDemuxer
	netProto   tcpip.NetworkProtocolNumber
	transProto tcpip.TransportProtocolNumber

	// endpoints stores the transport endpoints in the order in which they
	// were bound. This is required for UDP SO_REUSEADDR.
	endpoints []TransportEndpoint
	flags     ports.FlagCounter
}

type endpointsByNIC struct {
	mu        sync.RWMutex
	endpoints map[tcpip.NICID]*multiPortEndpoint
	// seed is a random secret for a jenkins hash.
	seed uint32
}

// transportEndpoints manages all endpoints of a given protocol. It has its own
// mutex so as to reduce interference between protocols.
type transportEndpoints struct {
	// mu protects all fields of the transportEndpoints.
	mu        sync.RWMutex
	endpoints map[TransportEndpointID]*endpointsByNIC
	// rawEndpoints contains endpoints for raw sockets, which receive all
	// traffic of a given protocol regardless of port.
	rawEndpoints []RawTransportEndpoint
}
