package dhcp6

import (
	"fmt"
	"encoding/binary"
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

func MakePacket(bs []byte, packetLength int) (*Packet, error) {
	options, err := MakeOptions(bs[4:packetLength])
	if err != nil {
		return nil, fmt.Errorf("packet has malformed options section: %s", err)
	}
	ret := &Packet{Type: MessageType(bs[0]), Options: options}
	copy(ret.TransactionID[:], bs[1:4])
	return ret, nil
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

func (p *Packet) BuildResponse(serverDuid []byte, addressPool AddressPool) *Packet {
	transactionId := p.TransactionID
	clientId := p.Options[OptClientId].Value
	iaNaId := p.Options[OptIaNa].Value[0:4]
	var clientArchType []byte
	o, exists := p.Options[OptClientArchType]; if exists {
		clientArchType = o.Value
	}
	switch p.Type {
	case MsgSolicit:
		return MakeMsgAdvertise(transactionId, serverDuid, clientId, iaNaId, clientArchType, addressPool)
	case MsgRequest:
		return MakeMsgReply(transactionId, serverDuid, clientId, iaNaId, clientArchType, addressPool)
	case MsgInformationRequest:
		return MakeMsgInformationRequestReply(transactionId, serverDuid, clientId, clientArchType)
	case MsgRelease:
		return MakeMsgReleaseReply(transactionId, serverDuid, clientId)
	default:
		return nil
	}
}

func MakeMsgAdvertise(transactionId [3]byte, serverDuid, clientId, iaId, clientArchType []byte, addressPool AddressPool) *Packet {
	ret_options := make(Options)
	ret_options.AddOption(MakeOption(OptClientId, clientId))
	association := addressPool.ReserveAddress(clientId, iaId)
	ret_options.AddOption(MakeIaNaOption(iaId, association.t1, association.t2,
		MakeIaAddrOption(association.ipAddress, 27000, 43200)))
	ret_options.AddOption(MakeOption(OptServerId, serverDuid))

	if 0x10 == binary.BigEndian.Uint16(clientArchType) { // HTTPClient
		ret_options.AddOption(MakeOption(OptVendorClass, []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116})) // HTTPClient
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte("http://[2001:db8:f00f:cafe::4]/bootx64.efi")))
	} else {
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte("http://[2001:db8:f00f:cafe::4]/script.ipxe")))
	}
	//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	//ret_options.AddOption(OptBootfileParam, []byte("http://")
	//ret.Options[OptPreference] = [][]byte("http://")

	return &Packet{Type: MsgAdvertise, TransactionID: transactionId, Options: ret_options}
}

// TODO: OptClientArchType may not be present

func MakeMsgReply(transactionId [3]byte, serverDuid, clientId, iaId, clientArchType []byte, addressPool AddressPool) *Packet {
	ret_options := make(Options)

	ret_options.AddOption(MakeOption(OptClientId, clientId))
	association := addressPool.ReserveAddress(clientId, iaId)
	ret_options.AddOption(MakeIaNaOption(iaId, association.t1, association.t2,
		MakeIaAddrOption(association.ipAddress, 27000, 43200)))
	ret_options.AddOption(MakeOption(OptServerId, serverDuid))
	//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	if 0x10 == binary.BigEndian.Uint16(clientArchType) { // HTTPClient
		ret_options.AddOption(MakeOption(OptVendorClass, []byte {0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116})) // HTTPClient
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte("http://[2001:db8:f00f:cafe::4]/bootx64.efi")))
	} else {
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte("http://[2001:db8:f00f:cafe::4]/script.ipxe")))
	}

	return &Packet{Type: MsgReply, TransactionID: transactionId, Options: ret_options}
}

func MakeMsgInformationRequestReply(transactionId [3]byte, serverDuid, clientId, clientArchType []byte) *Packet {
	ret_options := make(Options)
	ret_options.AddOption(MakeOption(OptClientId, clientId))
	ret_options.AddOption(MakeOption(OptServerId, serverDuid))
	//	ret_options.AddOption(OptRecursiveDns, net.ParseIP("2001:db8:f00f:cafe::1"))
	if 0x10 == binary.BigEndian.Uint16(clientArchType) { // HTTPClient
		ret_options.AddOption(MakeOption(OptVendorClass, []byte{0, 0, 0, 0, 0, 10, 72, 84, 84, 80, 67, 108, 105, 101, 110, 116})) // HTTPClient
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte("http://[2001:db8:f00f:cafe::4]/bootx64.efi")))
	} else {
		ret_options.AddOption(MakeOption(OptBootfileUrl, []byte("http://[2001:db8:f00f:cafe::4]/script.ipxe")))
	}

	return &Packet{Type: MsgReply, TransactionID: transactionId, Options: ret_options}
}

func MakeMsgReleaseReply(transactionId [3]byte, serverDuid, clientId []byte) *Packet {
	ret_options := make(Options)

	ret_options.AddOption(MakeOption(OptClientId, clientId))
	ret_options.AddOption(MakeOption(OptServerId, serverDuid))
	v := make([]byte, 19, 19)
	copy(v[2:], []byte("Release received."))
	ret_options.AddOption(MakeOption(OptStatusCode, v))

	return &Packet{Type: MsgReply, TransactionID: transactionId, Options: ret_options}
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