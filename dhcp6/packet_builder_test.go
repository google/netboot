package dhcp6

import (
	"testing"
	"encoding/binary"
	"net"
	"fmt"
)

func TestMakeMsgAdvertise(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	expectedInterfaceId := []byte("id-1")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IpAddress: expectedIp, InterfaceId: expectedInterfaceId}

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgAdvertise(transactionId, expectedClientId, 0x11, []*IdentityAssociation{identityAssociation},
		expectedBootFileUrl, nil)

	if msg.Type != MsgAdvertise {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options.ClientId()
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption)
	}

	serverIdOption := msg.Options.ServerId()
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption)
	}

	bootfileUrlOption := msg.Options[OptBootfileUrl][0]
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}

	iaNaOption := msg.Options.IaNaIds()
	if len(iaNaOption) == 0 {
		t.Fatalf("interface non-temporary association option should be present")
	}

	preferenceOption := msg.Options[OptPreference]
	if preferenceOption != nil {
		t.Fatalf("Preference option shouldn't be set")
	}
}

func TestShouldSetPreferenceOptionWhenSpecified(t *testing.T) {
	identityAssociation := &IdentityAssociation{IpAddress: net.ParseIP("2001:db8:f00f:cafe::1"), InterfaceId: []byte("id-1")}

	builder := MakePacketBuilder([]byte("serverid"), 90, 100)

	expectedPreference := []byte{128}
	msg := builder.MakeMsgAdvertise([3]byte{'t', 'i', 'd'}, []byte("clientid"), 0x11,
		[]*IdentityAssociation{identityAssociation}, []byte("http://bootfileurl"), expectedPreference)

	preferenceOption := msg.Options[OptPreference]
	if preferenceOption == nil {
		t.Fatalf("Preference option should be set")
	}
	if string(expectedPreference) != string(preferenceOption[0].Value) {
		t.Fatalf("Expected preference value %d, got %d", expectedPreference, preferenceOption[0].Value)
	}
}

func TestMakeMsgAdvertiseWithHttpClientArch(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IpAddress: expectedIp, InterfaceId: []byte("id-1")}

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgAdvertise(transactionId, expectedClientId, 0x10, []*IdentityAssociation{identityAssociation},
		expectedBootFileUrl, nil)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}
	bootfileUrlOption := msg.Options.BootfileUrl()
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption) {
		t.Fatalf("Expected bootfile URL %s, got %s", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeNoAddrsAvailable(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedMessage := "Boom!"

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgAdvertiseWithNoAddrsAvailable(transactionId, expectedClientId, fmt.Errorf(expectedMessage))

	if msg.Type != MsgAdvertise {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options.ClientId()
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption)
	}

	serverIdOption := msg.Options.ServerId()
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption)
	}

	_, exists := msg.Options[OptStatusCode]; if !exists {
		t.Fatalf("Expected status code option to be present")
	}
	statusCodeOption := msg.Options[OptStatusCode][0].Value
	if binary.BigEndian.Uint16(statusCodeOption[0:2]) != uint16(2) {
		t.Fatalf("Expected status code 2, got %d", binary.BigEndian.Uint16(statusCodeOption[0:2]))
	}
	if string(statusCodeOption[2:]) != expectedMessage {
		t.Fatalf("Expected message %s, got %s", expectedMessage, string(statusCodeOption[2:]))
	}
}

func TestMakeMsgReply(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IpAddress: expectedIp, InterfaceId: []byte("id-1")}

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgReply(transactionId, expectedClientId, 0x11, []*IdentityAssociation{identityAssociation},
		make([][]byte, 0), expectedBootFileUrl, nil)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options.ClientId()
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption)
	}
	if len(expectedClientId) != len(clientIdOption) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption))
	}

	serverIdOption := msg.Options.ServerId()
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption)
	}
	if len(expectedServerId) != len(serverIdOption) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption))
	}

	bootfileUrlOption := msg.Options.BootfileUrl()
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption) {
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
	identityAssociation := &IdentityAssociation{IpAddress: expectedIp, InterfaceId: []byte("id-1")}

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgReply(transactionId, expectedClientId, 0x10,
		[]*IdentityAssociation{identityAssociation}, make([][]byte, 0),
		expectedBootFileUrl, nil)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}

	bootfileUrlOption := msg.Options.BootfileUrl()
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgReplyWithNoAddrsAvailable(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileUrl := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IpAddress: expectedIp, InterfaceId: []byte("id-1")}
	expectedErrorMessage := "Boom!"

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgReply(transactionId, expectedClientId, 0x10,
		[]*IdentityAssociation{identityAssociation}, [][]byte{[]byte("id-2")}, expectedBootFileUrl,
		fmt.Errorf(expectedErrorMessage))

	iaNaOption := msg.Options[OptIaNa]
	if iaNaOption == nil {
		t.Fatalf("interface non-temporary association options should be present")
	}
	if (len(iaNaOption)) != 2 {
		t.Fatalf("Expected 2 identity associations, got %d", len(iaNaOption))
	}
	var okIaNaOption, failedIaNaOption []byte
	if string(iaNaOption[0].Value[0:4]) == string("id-1") {
		okIaNaOption = iaNaOption[0].Value
		failedIaNaOption = iaNaOption[1].Value
	} else {
		okIaNaOption = iaNaOption[1].Value
		failedIaNaOption = iaNaOption[0].Value
	}

	possiblyIaAddrOption, err := UnmarshalOption(okIaNaOption[12:])
	if err != nil {
		t.Fatalf("Failed to unmarshal IaNa options: %s", err)
	}
	if possiblyIaAddrOption.Id != OptIaAddr {
		t.Fatalf("Expected option 5 (ia address), got %d", possiblyIaAddrOption.Id)
	}

	possiblyStatusOption, err := UnmarshalOption(failedIaNaOption[12:])
	if err != nil {
		t.Fatalf("Failed to unmarshal IaNa options: %s", err)
	}
	if possiblyStatusOption.Id != OptStatusCode {
		t.Fatalf("Expected option 13 (status code), got %d", possiblyStatusOption.Id)
	}
	if binary.BigEndian.Uint16(possiblyStatusOption.Value[0:2]) != uint16(2) {
		t.Fatalf("Expected status code 2, got %d", binary.BigEndian.Uint16(possiblyStatusOption.Value[0:2]))
	}
	if string(possiblyStatusOption.Value[2:]) != expectedErrorMessage {
		t.Fatalf("Expected message %s, got %s", expectedErrorMessage, string(possiblyStatusOption.Value[2:]))
	}
}

func TestMakeMsgInformationRequestReply(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgInformationRequestReply(transactionId, expectedClientId, 0x11, expectedBootFileUrl)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options.ClientId()
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption)
	}
	if len(expectedClientId) != len(clientIdOption) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption))
	}

	serverIdOption := msg.Options.ServerId()
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption)
	}
	if len(expectedServerId) != len(serverIdOption) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption))
	}

	bootfileUrlOption := msg.Options.BootfileUrl()
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgInformationRequestReplyWithHttpClientArch(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}
	expectedBootFileUrl := []byte("http://bootfileurl")

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgInformationRequestReply(transactionId, expectedClientId, 0x10, expectedBootFileUrl)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}

	bootfileUrlOption := msg.Options.BootfileUrl()
	if bootfileUrlOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileUrl) != string(bootfileUrlOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileUrl, bootfileUrlOption)
	}
}

func TestMakeMsgReleaseReply(t *testing.T) {
	expectedClientId := []byte("clientid")
	expectedServerId := []byte("serverid")
	transactionId := [3]byte{'1', '2', '3'}

	builder := MakePacketBuilder(expectedServerId, 90, 100)

	msg := builder.MakeMsgReleaseReply(transactionId, expectedClientId)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionId != msg.TransactionID {
		t.Fatalf("Expected transaction Id %v, got %v", transactionId, msg.TransactionID)
	}

	clientIdOption := msg.Options.ClientId()
	if clientIdOption == nil {
		t.Fatalf("Client Id option should be present")
	}
	if string(expectedClientId) != string(clientIdOption) {
		t.Fatalf("Expected client id %v, got %v", expectedClientId, clientIdOption)
	}
	if len(expectedClientId) != len(clientIdOption) {
		t.Fatalf("Expected client id length of %d, got %d", len(expectedClientId), len(clientIdOption))
	}

	serverIdOption := msg.Options.ServerId()
	if serverIdOption == nil {
		t.Fatalf("Server Id option should be present")
	}
	if string(expectedServerId) != string(serverIdOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientId, serverIdOption)
	}
	if len(expectedServerId) != len(serverIdOption) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientId), len(serverIdOption))
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
