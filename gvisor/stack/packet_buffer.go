package stack

import (
	"github.com/icexin/eggos/gvisor/tcpip"
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/sync"
)

// PacketBufferOptions specifies options for PacketBuffer creation.
type PacketBufferOptions struct {
	// ReserveHeaderBytes is the number of bytes to reserve for headers. Total
	// number of bytes pushed onto the headers must not exceed this value.
	ReserveHeaderBytes int

	// Data is the initial unparsed data for the new packet. If set, it will be
	// owned by the new packet.
	Data tcpipbuffer.VectorisedView

	// IsForwardedPacket identifies that the PacketBuffer being created is for a
	// forwarded packet.
	IsForwardedPacket bool
}

// A PacketBuffer contains all the data of a network packet.
//
// As a PacketBuffer traverses up the stack, it may be necessary to pass it to
// multiple endpoints.
//
// The whole packet is expected to be a series of bytes in the following order:
// LinkHeader, NetworkHeader, TransportHeader, and Data. Any of them can be
// empty. Use of PacketBuffer in any other order is unsupported.
//
// PacketBuffer must be created with NewPacketBuffer.
//
// Internal structure: A PacketBuffer holds a pointer to buffer.Buffer, which
// exposes a logically-contiguous byte storage. The underlying storage structure
// is abstracted out, and should not be a concern here for most of the time.
//
// |- reserved ->|
//
//	|--->| consumed (incoming)
//
// 0             V    V
// +--------+----+----+--------------------+
// |        |    |    | current data ...   | (buf)
// +--------+----+----+--------------------+
//
//	^    |
//	|<---| pushed (outgoing)
//
// When a PacketBuffer is created, a `reserved` header region can be specified,
// which stack pushes headers in this region for an outgoing packet. There could
// be no such region for an incoming packet, and `reserved` is 0. The value of
// `reserved` never changes in the entire lifetime of the packet.
//
// Outgoing Packet: When a header is pushed, `pushed` gets incremented by the
// pushed length, and the current value is stored for each header. PacketBuffer
// substracts this value from `reserved` to compute the starting offset of each
// header in `buf`.
//
// Incoming Packet: When a header is consumed (a.k.a. parsed), the current
// `consumed` value is stored for each header, and it gets incremented by the
// consumed length. PacketBuffer adds this value to `reserved` to compute the
// starting offset of each header in `buf`.
type PacketBuffer struct {
	_ sync.NoCopy

	// PacketBufferEntry is used to build an intrusive list of
	// PacketBuffers.
	PacketBufferEntry

	// buf is the underlying buffer for the packet. See struct level docs for
	// details.
	buf      *buffer.Buffer
	reserved int
	pushed   int
	consumed int

	// headers stores metadata about each header.
	headers [numHeaderType]headerInfo

	// NetworkProtocolNumber is only valid when NetworkHeader().View().IsEmpty()
	// returns false.
	// TODO(gvisor.dev/issue/3574): Remove the separately passed protocol
	// numbers in registration APIs that take a PacketBuffer.
	NetworkProtocolNumber tcpip.NetworkProtocolNumber

	// TransportProtocol is only valid if it is non zero.
	// TODO(gvisor.dev/issue/3810): This and the network protocol number should
	// be moved into the headerinfo. This should resolve the validity issue.
	TransportProtocolNumber tcpip.TransportProtocolNumber

	// Hash is the transport layer hash of this packet. A value of zero
	// indicates no valid hash has been set.
	Hash uint32

	// Owner is implemented by task to get the uid and gid.
	// Only set for locally generated packets.
	Owner tcpip.PacketOwner

	// The following fields are only set by the qdisc layer when the packet
	// is added to a queue.
	EgressRoute RouteInfo
	GSOOptions  GSO

	// NatDone indicates if the packet has been manipulated as per NAT
	// iptables rule.
	NatDone bool

	// PktType indicates the SockAddrLink.PacketType of the packet as defined in
	// https://www.man7.org/linux/man-pages/man7/packet.7.html.
	PktType tcpip.PacketType

	// NICID is the ID of the last interface the network packet was handled at.
	NICID tcpip.NICID

	// RXTransportChecksumValidated indicates that transport checksum verification
	// may be safely skipped.
	RXTransportChecksumValidated bool

	// NetworkPacketInfo holds an incoming packet's network-layer information.
	NetworkPacketInfo NetworkPacketInfo
}

// Views returns the underlying storage of the whole packet.
func (pk *PacketBuffer) Views() []tcpipbuffer.View {
	var views []tcpipbuffer.View
	offset := pk.headerOffset()
	pk.buf.SubApply(offset, int(pk.buf.Size())-offset, func(v []byte) {
		views = append(views, v)
	})
	return views
}

func (pk *PacketBuffer) headerOffset() int {
	return pk.reserved - pk.pushed
}
