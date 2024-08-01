package stack

import (
	"sync"

	"github.com/icexin/eggos/gvisor/tcpip"
)

type Hook uint

const (
	// Prerouting happens before a packet is routed to applications or to
	// be forwarded.
	Prerouting Hook = iota

	// Input happens before a packet reaches an application.
	Input

	// Forward happens once it's decided that a packet should be forwarded
	// to another host.
	Forward

	// Output happens after a packet is written by an application to be
	// sent out.
	Output

	// Postrouting happens just before a packet goes out on the wire.
	Postrouting

	// NumHooks is the total number of hooks.
	NumHooks
)

// A Matcher is the interface for matching packets.
type Matcher interface {
	// Match returns whether the packet matches and whether the packet
	// should be "hotdropped", i.e. dropped immediately. This is usually
	// used for suspicious packets.
	//
	// Precondition: packet.NetworkHeader is set.
	Match(hook Hook, packet *PacketBuffer, inputInterfaceName, outputInterfaceName string) (matches bool, hotdrop bool)
}

// A Target is the interface for taking an action for a packet.
type Target interface {
	// Action takes an action on the packet and returns a verdict on how
	// traversal should (or should not) continue. If the return value is
	// Jump, it also returns the index of the rule to jump to.
	Action(*PacketBuffer, *ConnTrack, Hook, *Route, tcpip.Address) (RuleVerdict, int)
}

// A Rule is a packet processing rule. It consists of two pieces. First it
// contains zero or more matchers, each of which is a specification of which
// packets this rule applies to. If there are no matchers in the rule, it
// applies to any packet.
//
// +stateify savable
type Rule struct {
	// Filter holds basic IP filtering fields common to every rule.
	Filter IPHeaderFilter

	// Matchers is the list of matchers for this rule.
	Matchers []Matcher

	// Target is the action to invoke if all the matchers match the packet.
	Target Target
}

// A Table defines a set of chains and hooks into the network stack.
//
// It is a list of Rules, entry points (BuiltinChains), and error handlers
// (Underflows). As packets traverse netstack, they hit hooks. When a packet
// hits a hook, iptables compares it to Rules starting from that hook's entry
// point. So if a packet hits the Input hook, we look up the corresponding
// entry point in BuiltinChains and jump to that point.
//
// If the Rule doesn't match the packet, iptables continues to the next Rule.
// If a Rule does match, it can issue a verdict on the packet (e.g. RuleAccept
// or RuleDrop) that causes the packet to stop traversing iptables. It can also
// jump to other rules or perform custom actions based on Rule.Target.
//
// Underflow Rules are invoked when a chain returns without reaching a verdict.
//
// +stateify savable
type Table struct {
	// Rules holds the rules that make up the table.
	Rules []Rule

	// BuiltinChains maps builtin chains to their entrypoint rule in Rules.
	BuiltinChains [NumHooks]int

	// Underflows maps builtin chains to their underflow rule in Rules
	// (i.e. the rule to execute if the chain returns without a verdict).
	Underflows [NumHooks]int
}

// IPHeaderFilter performs basic IP header matching common to every rule.
//
// +stateify savable
type IPHeaderFilter struct {
	// Protocol matches the transport protocol.
	Protocol tcpip.TransportProtocolNumber

	// CheckProtocol determines whether the Protocol field should be
	// checked during matching.
	CheckProtocol bool

	// Dst matches the destination IP address.
	Dst tcpip.Address

	// DstMask masks bits of the destination IP address when comparing with
	// Dst.
	DstMask tcpip.Address

	// DstInvert inverts the meaning of the destination IP check, i.e. when
	// true the filter will match packets that fail the destination
	// comparison.
	DstInvert bool

	// Src matches the source IP address.
	Src tcpip.Address

	// SrcMask masks bits of the source IP address when comparing with Src.
	SrcMask tcpip.Address

	// SrcInvert inverts the meaning of the source IP check, i.e. when true the
	// filter will match packets that fail the source comparison.
	SrcInvert bool

	// InputInterface matches the name of the incoming interface for the packet.
	InputInterface string

	// InputInterfaceMask masks the characters of the interface name when
	// comparing with InputInterface.
	InputInterfaceMask string

	// InputInterfaceInvert inverts the meaning of incoming interface check,
	// i.e. when true the filter will match packets that fail the incoming
	// interface comparison.
	InputInterfaceInvert bool

	// OutputInterface matches the name of the outgoing interface for the packet.
	OutputInterface string

	// OutputInterfaceMask masks the characters of the interface name when
	// comparing with OutputInterface.
	OutputInterfaceMask string

	// OutputInterfaceInvert inverts the meaning of outgoing interface check,
	// i.e. when true the filter will match packets that fail the outgoing
	// interface comparison.
	OutputInterfaceInvert bool
}

// IPTables holds all the tables for a netstack.
//
// +stateify savable
type IPTables struct {
	// mu protects v4Tables, v6Tables, and modified.
	mu sync.RWMutex
	// v4Tables and v6tables map tableIDs to tables. They hold builtin
	// tables only, not user tables. mu must be locked for accessing.
	v4Tables [NumTables]Table
	v6Tables [NumTables]Table
	// modified is whether tables have been modified at least once. It is
	// used to elide the iptables performance overhead for workloads that
	// don't utilize iptables.
	modified bool

	// priorities maps each hook to a list of table names. The order of the
	// list is the order in which each table should be visited for that
	// hook. It is immutable.
	priorities [NumHooks][]TableID

	connections ConnTrack

	// reaperDone can be signaled to stop the reaper goroutine.
	reaperDone chan struct{}
}
