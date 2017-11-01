package dhcp6

import (
	"testing"
	"encoding/binary"
	"net"
	"fmt"
)

func TestMakeMsgAdvertise(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	expectedInterfaceID := []byte("id-1")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	expectedDNSServerIP := net.ParseIP("2001:db8:f00f:cafe::99")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: expectedInterfaceID}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgAdvertise(transactionID, expectedServerID, expectedClientID, 0x11,
		[]*IdentityAssociation{identityAssociation}, expectedBootFileURL, nil, []net.IP{expectedDNSServerIP})

	if msg.Type != MsgAdvertise {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionID != msg.TransactionID {
		t.Fatalf("Expected transaction ID %v, got %v", transactionID, msg.TransactionID)
	}

	clientIDOption := msg.Options.ClientID()
	if clientIDOption == nil {
		t.Fatalf("Client ID option should be present")
	}
	if string(expectedClientID) != string(clientIDOption) {
		t.Fatalf("Expected Client id %v, got %v", expectedClientID, clientIDOption)
	}

	serverIDOption := msg.Options.ServerID()
	if serverIDOption == nil {
		t.Fatalf("Server ID option should be present")
	}
	if string(expectedServerID) != string(serverIDOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientID, serverIDOption)
	}

	bootfileURLOption := msg.Options[OptBootfileURL][0]
	if bootfileURLOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileURL) != string(bootfileURLOption.Value) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileURL, bootfileURLOption)
	}

	iaNaOption := msg.Options.IaNaIDs()
	if len(iaNaOption) == 0 {
		t.Fatalf("interface non-temporary association option should be present")
	}

	preferenceOption := msg.Options[OptPreference]
	if preferenceOption != nil {
		t.Fatalf("Preference option shouldn't be set")
	}

	dnsServersOption := msg.Options[OptRecursiveDNS]
	if dnsServersOption == nil {
		t.Fatalf("DNS servers option should be set")
	}
	if string(dnsServersOption[0].Value) != string(expectedDNSServerIP) {
		t.Fatalf("Expected dns server %v, got %v", expectedDNSServerIP, net.IP(dnsServersOption[0].Value))
	}
}

func TestMakeMsgAdvertiseShouldSkipDnsServersIfNoneConfigured(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	expectedInterfaceID := []byte("id-1")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: expectedInterfaceID}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgAdvertise(transactionID, expectedServerID, expectedClientID, 0x11,
		[]*IdentityAssociation{identityAssociation}, expectedBootFileURL, nil, []net.IP{})

	_, exists := msg.Options[OptRecursiveDNS]; if exists {
		t.Fatalf("DNS servers option should not be set")
	}
}

func TestShouldSetPreferenceOptionWhenSpecified(t *testing.T) {
	identityAssociation := &IdentityAssociation{IPAddress: net.ParseIP("2001:db8:f00f:cafe::1"), InterfaceID: []byte("id-1")}

	builder := MakePacketBuilder(90, 100)

	expectedPreference := []byte{128}
	msg := builder.MakeMsgAdvertise([3]byte{'t', 'i', 'd'}, []byte("serverid"), []byte("clientid"), 0x11,
		[]*IdentityAssociation{identityAssociation}, []byte("http://bootfileurl"), expectedPreference, []net.IP{})

	preferenceOption := msg.Options[OptPreference]
	if preferenceOption == nil {
		t.Fatalf("Preference option should be set")
	}
	if string(expectedPreference) != string(preferenceOption[0].Value) {
		t.Fatalf("Expected preference value %d, got %d", expectedPreference, preferenceOption[0].Value)
	}
}

func TestMakeMsgAdvertiseWithHttpClientArch(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: []byte("id-1")}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgAdvertise(transactionID, expectedServerID, expectedClientID, 0x10,
		[]*IdentityAssociation{identityAssociation}, expectedBootFileURL, nil, []net.IP{})

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}
	bootfileURLOption := msg.Options.BootFileURL()
	if bootfileURLOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileURL) != string(bootfileURLOption) {
		t.Fatalf("Expected bootfile URL %s, got %s", expectedBootFileURL, bootfileURLOption)
	}
}

func TestMakeNoAddrsAvailable(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedMessage := "Boom!"

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgAdvertiseWithNoAddrsAvailable(transactionID, expectedServerID, expectedClientID, fmt.Errorf(expectedMessage))

	if msg.Type != MsgAdvertise {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionID != msg.TransactionID {
		t.Fatalf("Expected transaction ID %v, got %v", transactionID, msg.TransactionID)
	}

	clientIDOption := msg.Options.ClientID()
	if clientIDOption == nil {
		t.Fatalf("Client ID option should be present")
	}
	if string(expectedClientID) != string(clientIDOption) {
		t.Fatalf("Expected Client id %v, got %v", expectedClientID, clientIDOption)
	}

	serverIDOption := msg.Options.ServerID()
	if serverIDOption == nil {
		t.Fatalf("Server ID option should be present")
	}
	if string(expectedServerID) != string(serverIDOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientID, serverIDOption)
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
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	expectedDNSServerIP := net.ParseIP("2001:db8:f00f:cafe::99")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: []byte("id-1")}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgReply(transactionID, expectedServerID, expectedClientID, 0x11,
		[]*IdentityAssociation{identityAssociation}, make([][]byte, 0), expectedBootFileURL, []net.IP{expectedDNSServerIP}, nil)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionID != msg.TransactionID {
		t.Fatalf("Expected transaction ID %v, got %v", transactionID, msg.TransactionID)
	}

	clientIDOption := msg.Options.ClientID()
	if clientIDOption == nil {
		t.Fatalf("Client ID option should be present")
	}
	if string(expectedClientID) != string(clientIDOption) {
		t.Fatalf("Expected Client id %v, got %v", expectedClientID, clientIDOption)
	}
	if len(expectedClientID) != len(clientIDOption) {
		t.Fatalf("Expected Client id length of %d, got %d", len(expectedClientID), len(clientIDOption))
	}

	serverIDOption := msg.Options.ServerID()
	if serverIDOption == nil {
		t.Fatalf("Server ID option should be present")
	}
	if string(expectedServerID) != string(serverIDOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientID, serverIDOption)
	}
	if len(expectedServerID) != len(serverIDOption) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientID), len(serverIDOption))
	}

	bootfileURLOption := msg.Options.BootFileURL()
	if bootfileURLOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileURL) != string(bootfileURLOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileURL, bootfileURLOption)
	}

	iaNaOption := msg.Options[OptIaNa]
	if iaNaOption == nil {
		t.Fatalf("interface non-temporary association option should be present")
	}

	dnsServersOption := msg.Options[OptRecursiveDNS]
	if dnsServersOption == nil {
		t.Fatalf("DNS servers option should be set")
	}
	if string(dnsServersOption[0].Value) != string(expectedDNSServerIP) {
		t.Fatalf("Expected dns server %v, got %v", expectedDNSServerIP, net.IP(dnsServersOption[0].Value))
	}
}

func TestMakeMsgReplyShouldSkipDnsServersIfNoneWereConfigured(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: []byte("id-1")}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgReply(transactionID, expectedServerID, expectedClientID, 0x11,
		[]*IdentityAssociation{identityAssociation}, make([][]byte, 0), expectedBootFileURL, []net.IP{}, nil)

	_, exists := msg.Options[OptRecursiveDNS]; if exists {
		t.Fatalf("Dns servers option shouldn't be present")
	}
}

func TestMakeMsgReplyWithHttpClientArch(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: []byte("id-1")}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgReply(transactionID, expectedServerID, expectedClientID, 0x10,
		[]*IdentityAssociation{identityAssociation}, make([][]byte, 0), expectedBootFileURL, []net.IP{}, nil)

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}

	bootfileURLOption := msg.Options.BootFileURL()
	if bootfileURLOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileURL) != string(bootfileURLOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileURL, bootfileURLOption)
	}
}

func TestMakeMsgReplyWithNoAddrsAvailable(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedIP := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedBootFileURL := []byte("http://bootfileurl")
	identityAssociation := &IdentityAssociation{IPAddress: expectedIP, InterfaceID: []byte("id-1")}
	expectedErrorMessage := "Boom!"

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgReply(transactionID, expectedServerID, expectedClientID, 0x10,
		[]*IdentityAssociation{identityAssociation}, [][]byte{[]byte("id-2")}, expectedBootFileURL, []net.IP{},
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
	if possiblyIaAddrOption.ID != OptIaAddr {
		t.Fatalf("Expected option 5 (ia address), got %d", possiblyIaAddrOption.ID)
	}

	possiblyStatusOption, err := UnmarshalOption(failedIaNaOption[12:])
	if err != nil {
		t.Fatalf("Failed to unmarshal IaNa options: %s", err)
	}
	if possiblyStatusOption.ID != OptStatusCode {
		t.Fatalf("Expected option 13 (status code), got %d", possiblyStatusOption.ID)
	}
	if binary.BigEndian.Uint16(possiblyStatusOption.Value[0:2]) != uint16(2) {
		t.Fatalf("Expected status code 2, got %d", binary.BigEndian.Uint16(possiblyStatusOption.Value[0:2]))
	}
	if string(possiblyStatusOption.Value[2:]) != expectedErrorMessage {
		t.Fatalf("Expected message %s, got %s", expectedErrorMessage, string(possiblyStatusOption.Value[2:]))
	}
}

func TestMakeMsgInformationRequestReply(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedBootFileURL := []byte("http://bootfileurl")
	expectedDNSServerIP := net.ParseIP("2001:db8:f00f:cafe::99")

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgInformationRequestReply(transactionID, expectedServerID, expectedClientID, 0x11,
		expectedBootFileURL, []net.IP{expectedDNSServerIP})

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionID != msg.TransactionID {
		t.Fatalf("Expected transaction ID %v, got %v", transactionID, msg.TransactionID)
	}

	clientIDOption := msg.Options.ClientID()
	if clientIDOption == nil {
		t.Fatalf("Client ID option should be present")
	}
	if string(expectedClientID) != string(clientIDOption) {
		t.Fatalf("Expected Client id %v, got %v", expectedClientID, clientIDOption)
	}
	if len(expectedClientID) != len(clientIDOption) {
		t.Fatalf("Expected Client id length of %d, got %d", len(expectedClientID), len(clientIDOption))
	}

	serverIDOption := msg.Options.ServerID()
	if serverIDOption == nil {
		t.Fatalf("Server ID option should be present")
	}
	if string(expectedServerID) != string(serverIDOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientID, serverIDOption)
	}
	if len(expectedServerID) != len(serverIDOption) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientID), len(serverIDOption))
	}

	bootfileURLOption := msg.Options.BootFileURL()
	if bootfileURLOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileURL) != string(bootfileURLOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileURL, bootfileURLOption)
	}

	dnsServersOption := msg.Options[OptRecursiveDNS]
	if dnsServersOption == nil {
		t.Fatalf("DNS servers option should be set")
	}
	if string(dnsServersOption[0].Value) != string(expectedDNSServerIP) {
		t.Fatalf("Expected dns server %v, got %v", expectedDNSServerIP, net.IP(dnsServersOption[0].Value))
	}
}

func TestMakeMsgInformationRequestReplyShouldSkipDnsServersIfNoneWereConfigured(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedBootFileURL := []byte("http://bootfileurl")

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgInformationRequestReply(transactionID, expectedServerID, expectedClientID, 0x11,
		expectedBootFileURL, []net.IP{})

	_, exists := msg.Options[OptRecursiveDNS]; if exists {
		t.Fatalf("Dns servers option shouldn't be present")
	}
}

func TestMakeMsgInformationRequestReplyWithHttpClientArch(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}
	expectedBootFileURL := []byte("http://bootfileurl")

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgInformationRequestReply(transactionID, expectedServerID, expectedClientID, 0x10,
		expectedBootFileURL, []net.IP{})

	vendorClassOption := msg.Options[OptVendorClass]
	if vendorClassOption == nil {
		t.Fatalf("Vendor class option should be present")
	}

	bootfileURLOption := msg.Options.BootFileURL()
	if bootfileURLOption == nil {
		t.Fatalf("Bootfile URL option should be present")
	}
	if string(expectedBootFileURL) != string(bootfileURLOption) {
		t.Fatalf("Expected bootfile URL %v, got %v", expectedBootFileURL, bootfileURLOption)
	}
}

func TestMakeMsgReleaseReply(t *testing.T) {
	expectedClientID := []byte("clientid")
	expectedServerID := []byte("serverid")
	transactionID := [3]byte{'1', '2', '3'}

	builder := MakePacketBuilder(90, 100)

	msg := builder.MakeMsgReleaseReply(transactionID, expectedServerID, expectedClientID)

	if msg.Type != MsgReply {
		t.Fatalf("Expected message type %d, got %d", MsgAdvertise, msg.Type)
	}
	if transactionID != msg.TransactionID {
		t.Fatalf("Expected transaction ID %v, got %v", transactionID, msg.TransactionID)
	}

	clientIDOption := msg.Options.ClientID()
	if clientIDOption == nil {
		t.Fatalf("Client ID option should be present")
	}
	if string(expectedClientID) != string(clientIDOption) {
		t.Fatalf("Expected Client id %v, got %v", expectedClientID, clientIDOption)
	}
	if len(expectedClientID) != len(clientIDOption) {
		t.Fatalf("Expected Client id length of %d, got %d", len(expectedClientID), len(clientIDOption))
	}

	serverIDOption := msg.Options.ServerID()
	if serverIDOption == nil {
		t.Fatalf("Server ID option should be present")
	}
	if string(expectedServerID) != string(serverIDOption) {
		t.Fatalf("Expected server id %v, got %v", expectedClientID, serverIDOption)
	}
	if len(expectedServerID) != len(serverIDOption) {
		t.Fatalf("Expected server id length of %d, got %d", len(expectedClientID), len(serverIDOption))
	}
}

func TestExtractLLAddressOrIdWithDUIDLLT(t *testing.T) {
	builder := &PacketBuilder{}
	expectedLLAddress := []byte{0xac, 0xbc, 0x32, 0xae, 0x86, 0x37}
	llAddress := builder.ExtractLLAddressOrID([]byte{0x0, 0x1, 0x0, 0x1, 0x1, 0x2, 0x3, 0x4, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37})
	if string(expectedLLAddress) != string(llAddress) {
		t.Fatalf("Expected ll address %x, got: %x", expectedLLAddress, llAddress)
	}
}

func TestExtractLLAddressOrIdWithDUIDEN(t *testing.T) {
	builder := &PacketBuilder{}
	expectedID := []byte{0x0, 0x1, 0x2, 0x3, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37}
	id := builder.ExtractLLAddressOrID([]byte{0x0, 0x2, 0x0, 0x1, 0x2, 0x3, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37})
	if string(expectedID) != string(id) {
		t.Fatalf("Expected id %x, got: %x", expectedID, id)
	}
}

func TestExtractLLAddressOrIdWithDUIDLL(t *testing.T) {
	builder := &PacketBuilder{}
	expectedLLAddress := []byte{0xac, 0xbc, 0x32, 0xae, 0x86, 0x37}
	llAddress := builder.ExtractLLAddressOrID([]byte{0x0, 0x3, 0x0, 0x1, 0xac, 0xbc, 0x32, 0xae, 0x86, 0x37})
	if string(expectedLLAddress) != string(llAddress) {
		t.Fatalf("Expected ll address %x, got: %x", expectedLLAddress, llAddress)
	}
}
