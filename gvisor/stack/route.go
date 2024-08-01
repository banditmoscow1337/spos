package stack

import (
	"sync"

	"github.com/icexin/eggos/gvisor/tcpip"
)

// Route represents a route through the networking stack to a given destination.
//
// It is safe to call Route's methods from multiple goroutines.
type Route struct {
	routeInfo routeInfo

	// localAddressNIC is the interface the address is associated with.
	// TODO(gvisor.dev/issue/4548): Remove this field once we can query the
	// address's assigned status without the NIC.
	localAddressNIC *nic

	mu struct {
		sync.RWMutex

		// localAddressEndpoint is the local address this route is associated with.
		localAddressEndpoint AssignableAddressEndpoint

		// remoteLinkAddress is the link-layer (MAC) address of the next hop in the
		// route.
		remoteLinkAddress tcpip.LinkAddress
	}

	// outgoingNIC is the interface this route uses to write packets.
	outgoingNIC *nic

	// linkRes is set if link address resolution is enabled for this protocol on
	// the route's NIC.
	linkRes *linkResolver
}
