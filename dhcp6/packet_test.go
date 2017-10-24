package dhcp6

import (
	"testing"
	"encoding/binary"
	"net"
)

func TestMakeMsgAdvertise(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgAdvertise(transactionId, expectedClientId, []byte("1234"), 0x11, expectedIp,
		expectedBootFileUrl, nil)

	if msg.Type != MsgAdvertise {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options[OptClientId]
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption.Value) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption.Value)
	}
	if len(expectedClientId) != len(clientIdOption.Value) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption.Value))
	}

	serverIdOption := msg.Options[OptServerId]
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption.Value) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption.Value)
	}
	if len(expectedServerId) != len(serverIdOption.Value) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption.Value))
	}

	bootfileUrlOption := msg.Options[OptBootfileUrl]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}

	iaNaOption := msg.Options[OptIaNa]
	if iaNaOption == nil {
		t.Fatalf("interface non-temporary association option should be present")
	}

	preferenceOption := msg.Options[OptPreference]
	if preferenceOption != nil {
		t.Fatalf("Preference option shouldn't be set")
	}
}

func TestShouldSetPreferenceOptionWhenSpecified(t *testing.T) {
	builder := MakePacketBuilder([]byte("serverid"), 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	expectedPreference := []byte{128}
	msg := builder.MakeMsgAdvertise([3]byte{'t', 'i', 'd'}, []byte("clientid"), []byte("1234"), 0x11,
		net.ParseIP("2001:db8:f00f:cafe::1"), []byte("http://bootfileurl"), expectedPreference)

	preferenceOption := msg.Options[OptPreference]
	if preferenceOption == nil {
		t.Fatalf("Preference option should be set")
	}
	if string(expectedPreference) != string(preferenceOption.Value) {
		t.Fatalf("Expected preference value %d, got %d", expectedPreference, preferenceOption.Value)
	}
}

func TestMakeMsgAdvertiseWithHttpClientArch(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgAdvertise(transactionId, expectedClientId, []byte("1234"), 0x10, expectedIp,
		expectedBootFileUrl, nil)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}
	bootfileUrlOption := msg.Options[OptBootfileUrl]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %s, got %s", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgReply(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgReply(transactionId, expectedClientId, []byte("1234"), 0x11, expectedIp, expectedBootFileUrl)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options[OptClientId]
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption.Value) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption.Value)
	}
	if len(expectedClientId) != len(clientIdOption.Value) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption.Value))
	}

	serverIdOption := msg.Options[OptServerId]
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption.Value) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption.Value)
	}
	if len(expectedServerId) != len(serverIdOption.Value) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption.Value))
	}

	bootfileUrlOption := msg.Options[OptBootfileUrl]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}

	iaNaOption := msg.Options[OptIaNa]
	if iaNaOption == nil {
		t.Fatalf("interface non-temporary association option should be present")
	}
}

func TestMakeMsgReplyWithHttpClientArch(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgReply(transactionId, expectedClientId, []byte("1234"), 0x10, expectedIp, expectedBootFileUrl)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}

	bootfileUrlOption := msg.Options[OptBootfileUrl]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgInformationRequestReply(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgInformationRequestReply(transactionId, expectedClientId, 0x11, expectedBootFileUrl)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options[OptClientId]
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption.Value) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption.Value)
	}
	if len(expectedClientId) != len(clientIdOption.Value) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption.Value))
	}

	serverIdOption := msg.Options[OptServerId]
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption.Value) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption.Value)
	}
	if len(expectedServerId) != len(serverIdOption.Value) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption.Value))
	}

	bootfileUrlOption := msg.Options[OptBootfileUrl]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgInformationRequestReplyWithHttpClientArch(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgInformationRequestReply(transactionId, expectedClientId, 0x10, expectedBootFileUrl)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}

	bootfileUrlOption := msg.Options[OptBootfileUrl]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgReleaseReply(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}

	builder := MakePacketBuilder(expectedServerId, 90, 100, nil,
		NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), 100))

	msg := builder.MakeMsgReleaseReply(transactionId, expectedClientId)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options[OptClientId]
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption.Value) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption.Value)
	}
	if len(expectedClientId) != len(clientIdOption.Value) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption.Value))
	}

	serverIdOption := msg.Options[OptServerId]
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption.Value) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption.Value)
	}
	if len(expectedServerId) != len(serverIdOption.Value) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption.Value))
	}
}

func TestExtractLLAddressOrIdWithDUIDLLT(t *testing.T) {
	builder := &PacketBuilder{}
	expectedLLAddress := []byte{0xac, 0xbc, 0x32, 0xae, 0x86, 0x37}
	llAddress := builder.ExtractLLAddressOrId([]byte{0x0, 0x1, 0x0, 0x1, 0x1, 0x2, 0x3, 0x4, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37})
	if string(expectedLLAddress) != string(llAddress) {
		t.Fatalf("Expected ll address %x, got: %x", expectedLLAddress, llAddress)
	}
}

func TestExtractLLAddressOrIdWithDUIDEN(t *testing.T) {
	builder := &PacketBuilder{}
	expectedId := []byte{0x0, 0x1, 0x2, 0x3, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37}
	id := builder.ExtractLLAddressOrId([]byte{0x0, 0x2, 0x0, 0x1, 0x2, 0x3, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37})
	if string(expectedId) != string(id) {
		t.Fatalf("Expected id %x, got: %x", expectedId, id)
	}
}

func TestExtractLLAddressOrIdWithDUIDLL(t *testing.T) {
	builder := &PacketBuilder{}
	expectedLLAddress := []byte{0xac, 0xbc, 0x32, 0xae, 0x86, 0x37}
	llAddress := builder.ExtractLLAddressOrId([]byte{0x0, 0x3, 0x0, 0x1, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37})
	if string(expectedLLAddress) != string(llAddress) {
		t.Fatalf("Expected ll address %x, got: %x", expectedLLAddress, llAddress)
	}
}

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

