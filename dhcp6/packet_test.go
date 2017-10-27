package dhcp6

import (
	"testing"
	"encoding/binary"
)

func TestShouldDiscardSolicitWithoutBootfileUrlOption(t *testing.T) {
	clientId := []byte("clientid")
	options := make(Options)
	options.AddOption(&Option{Id: OptClientId, Length: uint16(len(clientId)), Value: clientId})
	solicit := &Packet{Type: MsgSolicit, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardSolicit(solicit); err == nil {
		t.Fatalf("Should discard solicit packet without bootfile url option, but didn't")
	}
}

func TestShouldDiscardSolicitWithoutClientIdOption(t *testing.T) {
	options := make(Options)
	options.AddOption(MakeOptionRequestOptions([]uint16{OptBootfileUrl}))
	solicit := &Packet{Type: MsgSolicit, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardSolicit(solicit); err == nil {
		t.Fatalf("Should discard solicit packet without client id option, but didn't")
	}
}

func TestShouldDiscardSolicitWithServerIdOption(t *testing.T) {
	serverId := []byte("serverid")
	clientId := []byte("clientid")
	options := make(Options)
	options.AddOption(MakeOptionRequestOptions([]uint16{OptBootfileUrl}))
	options.AddOption(&Option{Id: OptClientId, Length: uint16(len(clientId)), Value: clientId})
	options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverId)), Value: serverId})
	solicit := &Packet{Type: MsgSolicit, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardSolicit(solicit); err == nil {
		t.Fatalf("Should discard solicit packet with server id option, but didn't")
	}
}

func TestShouldDiscardRequestWithoutBootfileUrlOption(t *testing.T) {
	serverId := []byte("serverid")
	clientId := []byte("clientid")
	options := make(Options)
	options.AddOption(&Option{Id: OptClientId, Length: uint16(len(clientId)), Value: clientId})
	options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverId)), Value: serverId})
	request := &Packet{Type: MsgRequest, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardRequest(request, serverId); err == nil {
		t.Fatalf("Should discard request packet without bootfile url option, but didn't")
	}
}

func TestShouldDiscardRequestWithoutClientIdOption(t *testing.T) {
	serverId := []byte("serverid")
	options := make(Options)
	options.AddOption(MakeOptionRequestOptions([]uint16{OptBootfileUrl}))
	options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverId)), Value: serverId})
	request := &Packet{Type: MsgRequest, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardRequest(request, serverId); err == nil {
		t.Fatalf("Should discard request packet without client id option, but didn't")
	}
}

func TestShouldDiscardRequestWithoutServerIdOption(t *testing.T) {
	clientId := []byte("clientid")
	options := make(Options)
	options.AddOption(MakeOptionRequestOptions([]uint16{OptBootfileUrl}))
	options.AddOption(&Option{Id: OptClientId, Length: uint16(len(clientId)), Value: clientId})
	request := &Packet{Type: MsgRequest, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardRequest(request, []byte("serverid")); err == nil {
		t.Fatalf("Should discard request packet with server id option, but didn't")
	}
}

func TestShouldDiscardRequestWithWrongServerId(t *testing.T) {
	clientId := []byte("clientid")
	serverId := []byte("serverid")
	options := make(Options)
	options.AddOption(MakeOptionRequestOptions([]uint16{OptBootfileUrl}))
	options.AddOption(&Option{Id: OptClientId, Length: uint16(len(clientId)), Value: clientId})
	options.AddOption(&Option{Id: OptServerId, Length: uint16(len(serverId)), Value: serverId})
	request := &Packet{Type: MsgRequest, TransactionID: [3]byte{'1', '2', '3'}, Options: options}

	if err := ShouldDiscardRequest(request, []byte("wrongid")); err == nil {
		t.Fatalf("Should discard request packet with wrong server id option, but didn't")
	}
}

func MakeOptionRequestOptions(options []uint16) *Option {
	value := make([]byte, len(options)*2)
	for i, option := range(options) {
		binary.BigEndian.PutUint16(value[i*2:], option)
	}

	return &Option{Id: OptOro, Length: uint16(len(options)*2), Value: value}
}

