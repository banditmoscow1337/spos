package stack

import (
	"time"

	"github.com/icexin/eggos/gvisor/tcpip"
)

// NeighborState defines the state of a NeighborEntry within the Neighbor
// Unreachability Detection state machine, as per RFC 4861 section 7.3.2 and
// RFC 7048.
type NeighborState uint8

// NeighborEntry describes a neighboring device in the local network.
type NeighborEntry struct {
	Addr      tcpip.Address
	LinkAddr  tcpip.LinkAddress
	State     NeighborState
	UpdatedAt time.Time
}
