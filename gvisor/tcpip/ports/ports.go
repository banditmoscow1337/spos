package ports

import (
	"sync"

	"github.com/icexin/eggos/gvisor/tcpip"
)

// PortManager manages allocating, reserving and releasing ports.
type PortManager struct {
	// mu protects allocatedPorts.
	// LOCK ORDERING: mu > ephemeralMu.
	mu sync.RWMutex
	// allocatedPorts is a nesting of maps that ultimately map Reservations
	// to FlagCounters describing whether the Reservation is valid and can
	// be reused.
	allocatedPorts map[portDescriptor]addrToDevice

	// ephemeralMu protects firstEphemeral and numEphemeral.
	ephemeralMu    sync.RWMutex
	firstEphemeral uint16
	numEphemeral   uint16

	// hint is used to pick ports ephemeral ports in a stable order for
	// a given port offset.
	//
	// hint must be accessed using the portHint/incPortHint helpers.
	// TODO(gvisor.dev/issue/940): S/R this field.
	hint uint32
}

type portDescriptor struct {
	network   tcpip.NetworkProtocolNumber
	transport tcpip.TransportProtocolNumber
	port      uint16
}

type destination struct {
	addr tcpip.Address
	port uint16
}

// destToCounter maps each destination to the FlagCounter that represents
// endpoints to that destination.
//
// destToCounter is never empty. When it has no elements, it is removed from
// the map that references it.
type destToCounter map[destination]FlagCounter

// deviceToDest maps NICs to destinations for which there are port reservations.
//
// deviceToDest is never empty. When it has no elements, it is removed from the
// map that references it.
type deviceToDest map[tcpip.NICID]destToCounter

// addrToDevice maps IP addresses to NICs that have port reservations.
type addrToDevice map[tcpip.Address]deviceToDest
