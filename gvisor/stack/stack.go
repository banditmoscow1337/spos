package stack

import (
	"io"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/icexin/eggos/gvisor/tcpip"
	"github.com/icexin/eggos/gvisor/tcpip/ports"
	"github.com/icexin/eggos/gvisor/waiter"
)

// PrimaryEndpointBehavior is an enumeration of an AddressEndpoint's primary
// behavior.
type PrimaryEndpointBehavior int

const (
	// CanBePrimaryEndpoint indicates the endpoint can be used as a primary
	// endpoint for new connections with no local address. This is the
	// default when calling NIC.AddAddress.
	CanBePrimaryEndpoint PrimaryEndpointBehavior = iota

	// FirstPrimaryEndpoint indicates the endpoint should be the first
	// primary endpoint considered. If there are multiple endpoints with
	// this behavior, they are ordered by recency.
	FirstPrimaryEndpoint

	// NeverPrimaryEndpoint indicates the endpoint should never be a
	// primary endpoint.
	NeverPrimaryEndpoint
)

// RawFactory produces endpoints for writing various types of raw packets.
type RawFactory interface {
	// NewUnassociatedEndpoint produces endpoints for writing packets not
	// associated with a particular transport protocol. Such endpoints can
	// be used to write arbitrary packets that include the network header.
	NewUnassociatedEndpoint(stack *Stack, netProto tcpip.NetworkProtocolNumber, transProto tcpip.TransportProtocolNumber, waiterQueue *waiter.Queue) (tcpip.Endpoint, tcpip.Error)

	// NewPacketEndpoint produces endpoints for reading and writing packets
	// that include network and (when cooked is false) link layer headers.
	NewPacketEndpoint(stack *Stack, cooked bool, netProto tcpip.NetworkProtocolNumber, waiterQueue *waiter.Queue) (tcpip.Endpoint, tcpip.Error)
}

type transportProtocolState struct {
	proto          TransportProtocol
	defaultHandler func(id TransportEndpointID, pkt *PacketBuffer) bool
}

// ResumableEndpoint is an endpoint that needs to be resumed after restore.
type ResumableEndpoint interface {
	// Resume resumes an endpoint after restore. This can be used to restart
	// background workers such as protocol goroutines. This must be called after
	// all indirect dependencies of the endpoint has been restored, which
	// generally implies at the end of the restore process.
	Resume(*Stack)
}

// Stack is a networking stack, with all supported protocols, NICs, and route
// table.
type Stack struct {
	transportProtocols map[tcpip.TransportProtocolNumber]*transportProtocolState
	networkProtocols   map[tcpip.NetworkProtocolNumber]NetworkProtocol

	// rawFactory creates raw endpoints. If nil, raw endpoints are
	// disabled. It is set during Stack creation and is immutable.
	rawFactory RawFactory

	demux *transportDemuxer

	//stats tcpip.Stats

	// LOCK ORDERING: mu > route.mu.
	route struct {
		mu struct {
			sync.RWMutex

			table []tcpip.Route
		}
	}

	mu                       sync.RWMutex
	nics                     map[tcpip.NICID]*nic
	defaultForwardingEnabled map[tcpip.NetworkProtocolNumber]struct{}

	// cleanupEndpointsMu protects cleanupEndpoints.
	cleanupEndpointsMu sync.Mutex
	cleanupEndpoints   map[TransportEndpoint]struct{}

	*ports.PortManager

	// If not nil, then any new endpoints will have this probe function
	// invoked everytime they receive a TCP segment.
	tcpProbeFunc atomic.Value // TCPProbeFunc

	// clock is used to generate user-visible times.
	clock tcpip.Clock

	// handleLocal allows non-loopback interfaces to loop packets.
	handleLocal bool

	// tables are the iptables packet filtering and manipulation rules.
	// TODO(gvisor.dev/issue/4595): S/R this field.
	tables *IPTables

	// resumableEndpoints is a list of endpoints that need to be resumed if the
	// stack is being restored.
	resumableEndpoints []ResumableEndpoint

	// icmpRateLimiter is a global rate limiter for all ICMP messages generated
	// by the stack.
	icmpRateLimiter *ICMPRateLimiter

	// seed is a one-time random value initialized at stack startup
	// and is used to seed the TCP port picking on active connections
	//
	// TODO(gvisor.dev/issue/940): S/R this field.
	seed uint32

	// nudConfigs is the default NUD configurations used by interfaces.
	nudConfigs NUDConfigurations

	// nudDisp is the NUD event dispatcher that is used to send the netstack
	// integrator NUD related events.
	nudDisp NUDDispatcher

	// uniqueIDGenerator is a generator of unique identifiers.
	uniqueIDGenerator UniqueID

	// randomGenerator is an injectable pseudo random generator that can be
	// used when a random number is required.
	randomGenerator *rand.Rand

	// secureRNG is a cryptographically secure random number generator.
	secureRNG io.Reader

	// sendBufferSize holds the min/default/max send buffer sizes for
	// endpoints other than TCP.
	sendBufferSize tcpip.SendBufferSizeOption

	// receiveBufferSize holds the min/default/max receive buffer sizes for
	// endpoints other than TCP.
	receiveBufferSize tcpip.ReceiveBufferSizeOption

	// tcpInvalidRateLimit is the maximal rate for sending duplicate
	// acknowledgements in response to incoming TCP packets that are for an existing
	// connection but that are invalid due to any of the following reasons:
	//
	//   a) out-of-window sequence number.
	//   b) out-of-window acknowledgement number.
	//   c) PAWS check failure (when implemented).
	//
	// This is required to prevent potential ACK loops.
	// Setting this to 0 will disable all rate limiting.
	tcpInvalidRateLimit time.Duration
}

// UniqueID is an abstract generator of unique identifiers.
type UniqueID interface {
	UniqueID() uint64
}

// UniqueID returns a unique identifier.
func (s *Stack) UniqueID() uint64 {
	return s.uniqueIDGenerator.UniqueID()
}

// RemoveAddress removes an existing network-layer address from the specified
// NIC.
func (s *Stack) RemoveAddress(id tcpip.NICID, addr tcpip.Address) tcpip.Error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if nic, ok := s.nics[id]; ok {
		return nic.removeAddress(addr)
	}

	return &tcpip.ErrUnknownNICID{}
}

// AddAddress adds a new network-layer address to the specified NIC.
func (s *Stack) AddAddress(id tcpip.NICID, protocol tcpip.NetworkProtocolNumber, addr tcpip.Address) tcpip.Error {
	return s.AddAddressWithOptions(id, protocol, addr, CanBePrimaryEndpoint)
}

// AddAddressWithOptions is the same as AddAddress, but allows you to specify
// whether the new endpoint can be primary or not.
func (s *Stack) AddAddressWithOptions(id tcpip.NICID, protocol tcpip.NetworkProtocolNumber, addr tcpip.Address, peb PrimaryEndpointBehavior) tcpip.Error {
	netProto, ok := s.networkProtocols[protocol]
	if !ok {
		return &tcpip.ErrUnknownProtocol{}
	}
	return s.AddProtocolAddressWithOptions(id, tcpip.ProtocolAddress{
		Protocol: protocol,
		AddressWithPrefix: tcpip.AddressWithPrefix{
			Address:   addr,
			PrefixLen: netProto.DefaultPrefixLen(),
		},
	}, peb)
}

// AddProtocolAddressWithOptions is the same as AddProtocolAddress, but allows
// you to specify whether the new endpoint can be primary or not.
func (s *Stack) AddProtocolAddressWithOptions(id tcpip.NICID, protocolAddress tcpip.ProtocolAddress, peb PrimaryEndpointBehavior) tcpip.Error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nic, ok := s.nics[id]
	if !ok {
		return &tcpip.ErrUnknownNICID{}
	}

	return nic.addAddress(protocolAddress, peb)
}

// NewEndpoint creates a new transport layer endpoint of the given protocol.
func (s *Stack) NewEndpoint(transport tcpip.TransportProtocolNumber, network tcpip.NetworkProtocolNumber, waiterQueue *waiter.Queue) (tcpip.Endpoint, tcpip.Error) {
	t, ok := s.transportProtocols[transport]
	if !ok {
		return nil, &tcpip.ErrUnknownProtocol{}
	}

	return t.proto.NewEndpoint(network, waiterQueue)
}
