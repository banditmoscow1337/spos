package stack

import (
	"sync"

	"github.com/icexin/eggos/gvisor/tcpip"
)

var _ NetworkInterface = (*nic)(nil)

// nic represents a "network interface card" to which the networking stack is
// attached.
type nic struct {
	LinkEndpoint

	stack   *Stack
	id      tcpip.NICID
	name    string
	context NICContext

	stats sharedStats

	// The network endpoints themselves may be modified by calling the interface's
	// methods, but the map reference and entries must be constant.
	networkEndpoints          map[tcpip.NetworkProtocolNumber]NetworkEndpoint
	linkAddrResolvers         map[tcpip.NetworkProtocolNumber]*linkResolver
	duplicateAddressDetectors map[tcpip.NetworkProtocolNumber]DuplicateAddressDetector

	// enabled is set to 1 when the NIC is enabled and 0 when it is disabled.
	//
	// Must be accessed using atomic operations.
	enabled uint32

	// linkResQueue holds packets that are waiting for link resolution to
	// complete.
	linkResQueue packetsPendingLinkResolution

	mu struct {
		sync.RWMutex
		spoofing    bool
		promiscuous bool
		// packetEPs is protected by mu, but the contained packetEndpointList are
		// not.
		packetEPs map[tcpip.NetworkProtocolNumber]*packetEndpointList
	}
}

// removeAddress removes an address from n.
func (n *nic) removeAddress(addr tcpip.Address) tcpip.Error {
	for _, ep := range n.networkEndpoints {
		addressableEndpoint, ok := ep.(AddressableEndpoint)
		if !ok {
			continue
		}

		switch err := addressableEndpoint.RemovePermanentAddress(addr); err.(type) {
		case *tcpip.ErrBadLocalAddress:
			continue
		default:
			return err
		}
	}

	return &tcpip.ErrBadLocalAddress{}
}

// addAddress adds a new address to n, so that it starts accepting packets
// targeted at the given address (and network protocol).
func (n *nic) addAddress(protocolAddress tcpip.ProtocolAddress, peb PrimaryEndpointBehavior) tcpip.Error {
	ep, ok := n.networkEndpoints[protocolAddress.Protocol]
	if !ok {
		return &tcpip.ErrUnknownProtocol{}
	}

	addressableEndpoint, ok := ep.(AddressableEndpoint)
	if !ok {
		return &tcpip.ErrNotSupported{}
	}

	addressEndpoint, err := addressableEndpoint.AddAndAcquirePermanentAddress(protocolAddress.AddressWithPrefix, peb, AddressConfigStatic, false /* deprecated */)
	if err == nil {
		// We have no need for the address endpoint.
		addressEndpoint.DecRef()
	}
	return err
}
