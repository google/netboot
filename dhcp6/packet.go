package dhcp6

import (
	"fmt"
	"bytes"
)

type MessageType uint8

const (
	MsgSolicit MessageType = iota + 1
	MsgAdvertise
	MsgRequest
	MsgConfirm
	MsgRenew
	MsgRebind
	MsgReply
	MsgRelease
	MsgDecline
	MsgReconfigure
	MsgInformationRequest
	MsgRelayForw
	MsgRelayRepl
)

type Packet struct {
	Type          MessageType
	TransactionID [3]byte
	Options       Options
}

type PacketBuilder struct {
	ServerDuid				[]byte
	PreferredLifetime 		uint32
	ValidLifetime			uint32
	BootFileUrlForHttpBoot	string
	BootFileUrlForIpxe		string
	Addresses				AddressPool
}

func MakePacket(bs []byte, packetLength int) (*Packet, error) {
	options, err := MakeOptions(bs[4:packetLength])
	if err != nil {
		return nil, fmt.Errorf("packet has malformed options section: %s", err)
	}
	ret := &Packet{Type: MessageType(bs[0]), Options: options}
	copy(ret.TransactionID[:], bs[1:4])
	return ret, nil
}

func MakePacketBuilder(serverDuid []byte, preferredLifetime, validLifetime uint32, httpBootFileUrl, ipxeBootFileUrl string,
	addressPool AddressPool) *PacketBuilder {
	return &PacketBuilder{ServerDuid: serverDuid, PreferredLifetime: preferredLifetime, ValidLifetime: validLifetime,
		BootFileUrlForHttpBoot: httpBootFileUrl, BootFileUrlForIpxe: ipxeBootFileUrl, Addresses: addressPool}
}

func (p *Packet) Marshal() ([]byte, error) {
	marshalled_options, err := p.Options.Marshal()
	if err != nil {
		return nil, fmt.Errorf("packet has malformed options section: %s", err)
	}

	ret := make([]byte, len(marshalled_options) + 4, len(marshalled_options) + 4)
	ret[0] = byte(p.Type)
	copy(ret[1:], p.TransactionID[:])
	copy(ret[4:], marshalled_options)

	return ret, nil
}

func (p *Packet) ShouldDiscard(serverDuid []byte) error {
	switch p.Type {
	case MsgSolicit:
		return ShouldDiscardSolicit(p)
	case MsgRequest:
		return ShouldDiscardRequest(p, serverDuid)
	case MsgInformationRequest:
		return ShouldDiscardInformationRequest(p, serverDuid)
	case MsgRelease:
		return nil // FIX ME!
	default:
		return fmt.Errorf("Unknown packet")
	}
}

func ShouldDiscardSolicit(p *Packet) error {
	options := p.Options
	if !options.RequestedBootFileUrlOption() {
		return fmt.Errorf("'Solicit' packet doesn't have file url option")
	}
	if !options.HasClientId() {
		return fmt.Errorf("'Solicit' packet has no client id option")
	}
	if options.HasServerId() {
		return fmt.Errorf("'Solicit' packet has server id option")
	}
	return nil
}

func ShouldDiscardRequest(p *Packet, serverDuid []byte) error {
	options := p.Options
	if !options.RequestedBootFileUrlOption() {
		return fmt.Errorf("'Request' packet doesn't have file url option")
	}
	if !options.HasClientId() {
		return fmt.Errorf("'Request' packet has no client id option")
	}
	if !options.HasServerId() {
		return fmt.Errorf("'Request' packet has no server id option")
	}
	if bytes.Compare(options[OptServerId].Value, serverDuid) != 0 {
		return fmt.Errorf("'Request' packet's server id option (%d) is different from ours (%d)", options[OptServerId].Value, serverDuid)
	}
	return nil
}

func ShouldDiscardInformationRequest(p *Packet, serverDuid []byte) error {
	options := p.Options
	if !options.RequestedBootFileUrlOption() {
		return fmt.Errorf("'Information-request' packet doesn't have boot file url option")
	}
	if options.HasIaNa() || options.HasIaTa() {
		return fmt.Errorf("'Information-request' packet has an IA option present")
	}
	if options.HasServerId() && (bytes.Compare(options[OptServerId].Value, serverDuid) != 0) {
		return fmt.Errorf("'Information-request' packet's server id option (%d) is different from ours (%d)", options[OptServerId].Value, serverDuid)
	}
	return nil
}

func (b *PacketBuilder) BuildResponse(in *Packet) *Packet {
	switch in.Type {
	case MsgSolicit:
		association := b.Addresses.ReserveAddress(in.Options.ClientId(), in.Options.IaNaId())
		return b.MakeMsgAdvertise(in.TransactionID, in.Options.ClientId(), in.Options.IaNaId(),
			in.Options.ClientArchType(), association.ipAddress)
	case MsgRequest:
		association := b.Addresses.ReserveAddress(in.Options.ClientId(), in.Options.IaNaId())
		return b.MakeMsgReply(in.TransactionID, in.Options.ClientId(), in.Options.IaNaId(),
			in.Options.ClientArchType(), association.ipAddress)
	case MsgInformationRequest:
		return b.MakeMsgInformationRequestReply(in.TransactionID, in.Options.ClientId(),
			in.Options.ClientArchType())
	case MsgRelease:
		b.Addresses.ReleaseAddress(in.Options.ClientId(), in.Options.IaNaId())
		return b.MakeMsgReleaseReply(in.TransactionID, in.Options.ClientId())
	default:
		return nil
	}
}

func (b *PacketBuilder) MakeMsgAdvertise(transactionId [3]byte, clientId, iaId []byte, clientArchType uint16, ipAddress []byte) *Packet {
	ret_options := make(Options)
	ret_options.AddOption(MakeOption(OptClientId, clientId))
	ret_options.AddOption(MakeIaNaOption(iaId, b.calculateT1(), b.calculateT2(),
		MakeIaAddrOption(ipAddress, b.PreferredLifetime, b.ValidLifetime)))
	ret_options.AddOption(MakeOption(OptServerId, b.ServerDuid))

	if 0x10 ==  clientArchType { // HTTPClient
		ret_options.AddOption(MakeOption(OptVendorClass, []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116})) // HTTPClient
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte(b.BootFileUrlForHttpBoot)))
	} else {
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte(b.BootFileUrlForIpxe)))
	}
	//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	//ret_options.AddOption(OptBootfileParam, []byte("http://")
	//ret.Options[OptPreference] = [][]byte("http://")

	return &Packet{Type: MsgAdvertise, TransactionID: transactionId, Options: ret_options}
}

func (b *PacketBuilder) MakeMsgReply(transactionId [3]byte, clientId, iaId []byte, clientArchType uint16, ipAddress []byte) *Packet {
	ret_options := make(Options)
	ret_options.AddOption(MakeOption(OptClientId, clientId))
	ret_options.AddOption(MakeIaNaOption(iaId, b.calculateT1(), b.calculateT2(),
		MakeIaAddrOption(ipAddress, b.PreferredLifetime, b.ValidLifetime)))
	ret_options.AddOption(MakeOption(OptServerId, b.ServerDuid))

	if 0x10 ==  clientArchType { // HTTPClient
		ret_options.AddOption(MakeOption(OptVendorClass, []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116})) // HTTPClient
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte(b.BootFileUrlForHttpBoot)))
	} else {
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte(b.BootFileUrlForIpxe)))
	}

	return &Packet{Type: MsgReply, TransactionID: transactionId, Options: ret_options}
}

func (b *PacketBuilder) MakeMsgInformationRequestReply(transactionId [3]byte, clientId []byte, clientArchType uint16) *Packet {
	ret_options := make(Options)
	ret_options.AddOption(MakeOption(OptClientId, clientId))
	ret_options.AddOption(MakeOption(OptServerId, b.ServerDuid))

	if 0x10 ==  clientArchType { // HTTPClient
		ret_options.AddOption(MakeOption(OptVendorClass, []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116})) // HTTPClient
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte(b.BootFileUrlForHttpBoot)))
	} else {
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte(b.BootFileUrlForIpxe)))
	}

	return &Packet{Type: MsgReply, TransactionID: transactionId, Options: ret_options}
}

func (b *PacketBuilder) MakeMsgReleaseReply(transactionId [3]byte, clientId []byte) *Packet {
	ret_options := make(Options)

	ret_options.AddOption(MakeOption(OptClientId, clientId))
	ret_options.AddOption(MakeOption(OptServerId, b.ServerDuid))
	v := make([]byte, 19, 19)
	copy(v[2:], []byte("Release received."))
	ret_options.AddOption(MakeOption(OptStatusCode, v))

	return &Packet{Type: MsgReply, TransactionID: transactionId, Options: ret_options}
}

func (b *PacketBuilder) calculateT1() uint32 {
	return b.PreferredLifetime / 2
}

func (b *PacketBuilder) calculateT2() uint32 {
	return (b.PreferredLifetime * 4)/5
}

