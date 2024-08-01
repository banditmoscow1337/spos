package header

import "github.com/icexin/eggos/gvisor/tcpip"

// IPv4 is an IPv4 header.
// Most of the methods of IPv4 access to the underlying slice without
// checking the boundaries and could panic because of 'index out of range'.
// Always call IsValid() to validate an instance of IPv4 before using other
// methods.
type IPv4 []byte

const (
	// IPv4MinimumSize is the minimum size of a valid IPv4 packet;
	// i.e. a packet header with no options.
	IPv4MinimumSize = 20

	// IPv4MaximumHeaderSize is the maximum size of an IPv4 header. Given
	// that there are only 4 bits (max 0xF (15)) to represent the header length
	// in 32-bit (4 byte) units, the header cannot exceed 15*4 = 60 bytes.
	IPv4MaximumHeaderSize = 60

	// IPv4MaximumOptionsSize is the largest size the IPv4 options can be.
	IPv4MaximumOptionsSize = IPv4MaximumHeaderSize - IPv4MinimumSize

	// IPv4MaximumPayloadSize is the maximum size of a valid IPv4 payload.
	//
	// Linux limits this to 65,515 octets (the max IP datagram size - the IPv4
	// header size). But RFC 791 section 3.2 discusses the design of the IPv4
	// fragment "allows 2**13 = 8192 fragments of 8 octets each for a total of
	// 65,536 octets. Note that this is consistent with the datagram total
	// length field (of course, the header is counted in the total length and not
	// in the fragments)."
	IPv4MaximumPayloadSize = 65536

	// MinIPFragmentPayloadSize is the minimum number of payload bytes that
	// the first fragment must carry when an IPv4 packet is fragmented.
	MinIPFragmentPayloadSize = 8

	// IPv4AddressSize is the size, in bytes, of an IPv4 address.
	IPv4AddressSize = 4

	// IPv4ProtocolNumber is IPv4's network protocol number.
	IPv4ProtocolNumber tcpip.NetworkProtocolNumber = 0x0800

	// IPv4Version is the version of the IPv4 protocol.
	IPv4Version = 4

	// IPv4AllSystems is the all systems IPv4 multicast address as per
	// IANA's IPv4 Multicast Address Space Registry. See
	// https://www.iana.org/assignments/multicast-addresses/multicast-addresses.xhtml.
	IPv4AllSystems tcpip.Address = "\xe0\x00\x00\x01"

	// IPv4Broadcast is the broadcast address of the IPv4 procotol.
	IPv4Broadcast tcpip.Address = "\xff\xff\xff\xff"

	// IPv4Any is the non-routable IPv4 "any" meta address.
	IPv4Any tcpip.Address = "\x00\x00\x00\x00"

	// IPv4AllRoutersGroup is a multicast address for all routers.
	IPv4AllRoutersGroup tcpip.Address = "\xe0\x00\x00\x02"

	// IPv4MinimumProcessableDatagramSize is the minimum size of an IP
	// packet that every IPv4 capable host must be able to
	// process/reassemble.
	IPv4MinimumProcessableDatagramSize = 576

	// IPv4MinimumMTU is the minimum MTU required by IPv4, per RFC 791,
	// section 3.2:
	//   Every internet module must be able to forward a datagram of 68 octets
	//   without further fragmentation.  This is because an internet header may be
	//   up to 60 octets, and the minimum fragment is 8 octets.
	IPv4MinimumMTU = 68
)
