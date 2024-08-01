package stack

import (
	"time"

	"github.com/icexin/eggos/gvisor/tcpip"
)

// NUDConfigurations is the NUD configurations for the netstack. This is used
// by the neighbor cache to operate the NUD state machine on each device in the
// local network.
type NUDConfigurations struct {
	// BaseReachableTime is the base duration for computing the random reachable
	// time.
	//
	// Reachable time is the duration for which a neighbor is considered
	// reachable after a positive reachability confirmation is received. It is a
	// function of uniformly distributed random value between minRandomFactor and
	// maxRandomFactor multiplied by baseReachableTime. Using a random component
	// eliminates the possibility that Neighbor Unreachability Detection messages
	// will synchronize with each other.
	//
	// After this time, a neighbor entry will transition from REACHABLE to STALE
	// state.
	//
	// Must be greater than 0.
	BaseReachableTime time.Duration

	// LearnBaseReachableTime enables learning BaseReachableTime during runtime
	// from the neighbor discovery protocol, if supported.
	//
	// TODO(gvisor.dev/issue/2240): Implement this NUD configuration option.
	LearnBaseReachableTime bool

	// MinRandomFactor is the minimum value of the random factor used for
	// computing reachable time.
	//
	// See BaseReachbleTime for more information on computing the reachable time.
	//
	// Must be greater than 0.
	MinRandomFactor float32

	// MaxRandomFactor is the maximum value of the random factor used for
	// computing reachabile time.
	//
	// See BaseReachbleTime for more information on computing the reachable time.
	//
	// Must be great than or equal to MinRandomFactor.
	MaxRandomFactor float32

	// RetransmitTimer is the duration between retransmission of reachability
	// probes in the PROBE state.
	RetransmitTimer time.Duration

	// LearnRetransmitTimer enables learning RetransmitTimer during runtime from
	// the neighbor discovery protocol, if supported.
	//
	// TODO(gvisor.dev/issue/2241): Implement this NUD configuration option.
	LearnRetransmitTimer bool

	// DelayFirstProbeTime is the duration to wait for a non-Neighbor-Discovery
	// related protocol to reconfirm reachability after entering the DELAY state.
	// After this time, a reachability probe will be sent and the entry will
	// transition to the PROBE state.
	//
	// Must be greater than 0.
	DelayFirstProbeTime time.Duration

	// MaxMulticastProbes is the number of reachability probes to send before
	// concluding negative reachability and deleting the neighbor entry from the
	// INCOMPLETE state.
	//
	// Must be greater than 0.
	MaxMulticastProbes uint32

	// MaxUnicastProbes is the number of reachability probes to send before
	// concluding retransmission from within the PROBE state should cease and
	// entry SHOULD be deleted.
	//
	// Must be greater than 0.
	MaxUnicastProbes uint32

	// MaxAnycastDelayTime is the time in which the stack SHOULD delay sending a
	// response for a random time between 0 and this time, if the target address
	// is an anycast address.
	//
	// TODO(gvisor.dev/issue/2242): Use this option when sending solicited
	// neighbor confirmations to anycast addresses and proxying neighbor
	// confirmations.
	MaxAnycastDelayTime time.Duration

	// MaxReachabilityConfirmations is the number of unsolicited reachability
	// confirmation messages a node MAY send to all-node multicast address when
	// it determines its link-layer address has changed.
	//
	// TODO(gvisor.dev/issue/2246): Discuss if implementation of this NUD
	// configuration option is necessary.
	MaxReachabilityConfirmations uint32
}

// NUDDispatcher is the interface integrators of netstack must implement to
// receive and handle NUD related events.
type NUDDispatcher interface {
	// OnNeighborAdded will be called when a new entry is added to a NIC's (with
	// ID nicID) neighbor table.
	//
	// This function is permitted to block indefinitely without interfering with
	// the stack's operation.
	//
	// May be called concurrently.
	OnNeighborAdded(tcpip.NICID, NeighborEntry)

	// OnNeighborChanged will be called when an entry in a NIC's (with ID nicID)
	// neighbor table changes state and/or link address.
	//
	// This function is permitted to block indefinitely without interfering with
	// the stack's operation.
	//
	// May be called concurrently.
	OnNeighborChanged(tcpip.NICID, NeighborEntry)

	// OnNeighborRemoved will be called when an entry is removed from a NIC's
	// (with ID nicID) neighbor table.
	//
	// This function is permitted to block indefinitely without interfering with
	// the stack's operation.
	//
	// May be called concurrently.
	OnNeighborRemoved(tcpip.NICID, NeighborEntry)
}
