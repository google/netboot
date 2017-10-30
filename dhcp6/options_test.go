package dhcp6

import (
	"testing"
	"encoding/binary"
	"net"
)

func TestMarshalOption(t *testing.T) {
	expectedUrl := []byte("http://blah")
	expectedLength := uint16(len(expectedUrl))
	opt := &Option{Id: OptBootfileUrl, Length: expectedLength, Value: expectedUrl}

	marshalled, err := opt.Marshal()
	if err != nil {
		t.Fatalf("Unexpected marshalling failure.")
	}

	if id := uint16(binary.BigEndian.Uint16(marshalled[0:2])); id != OptBootfileUrl {
		t.Fatalf("Expected optionId %d, got: %d", OptBootfileUrl, id)
	}
	if l := uint16(binary.BigEndian.Uint16(marshalled[2:4])); l != expectedLength {
		t.Fatalf("Expected length of %d, got: %d", expectedLength, l)
	}
	if url := marshalled[4:]; string(url) != string(expectedUrl) {
		t.Fatalf("Expected %s, got %s", expectedUrl, url)
	}
}

func TestMakeIaAddrOption(t *testing.T) {
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::99")
	var expectedPreferredLifetime, expectedValidLifetime uint32 = 27000, 43200
	iaAddrOption := MakeIaAddrOption(expectedIp, expectedPreferredLifetime, expectedValidLifetime)

	if iaAddrOption.Id != OptIaAddr {
		t.Fatalf("Expected option id %d, got %d", OptIaAddr, iaAddrOption.Id)
	}
	if iaAddrOption.Length != 24 {
		t.Fatalf("Expected length 24 bytes, got %d", iaAddrOption.Length)
	}
	if string(iaAddrOption.Value[0:16]) != string(expectedIp) {
		t.Fatalf("Expected address %v, got %v", expectedIp, iaAddrOption.Value[0:16])
	}
	if preferredLifetime := uint32(binary.BigEndian.Uint32(iaAddrOption.Value[16:20])); preferredLifetime != expectedPreferredLifetime {
		t.Fatalf("Expected preferred lifetime of %d, got %d", expectedPreferredLifetime, preferredLifetime)
	}
	if validLifetime := uint32(binary.BigEndian.Uint32(iaAddrOption.Value[20:24])); validLifetime != expectedValidLifetime {
		t.Fatalf("Expected valid lifetime of %d, got %d", expectedValidLifetime, validLifetime)
	}
}

func TestMakeIaNaOption(t *testing.T) {
	iaAddrOption := MakeIaAddrOption(net.ParseIP("2001:db8:f00f:cafe::100"), 100, 200)
	expectedSerializedIaAddrOption, err := iaAddrOption.Marshal()
	if err != nil {
		t.Fatalf("Unexpected serialization error: %s", err)
	}
	expectedId := []byte("1234")
	var expectedT1, expectedT2 uint32 = 100, 200

	iaNaOption := MakeIaNaOption(expectedId, expectedT1, expectedT2, iaAddrOption)

	if iaNaOption.Id != OptIaNa {
		t.Fatalf("Expected optionId %d, got %d", OptIaNa, iaNaOption.Id)
	}
	if string(iaNaOption.Value[0:4]) != string(expectedId) {
		t.Fatalf("Expected id %s, got %s", expectedId, string(iaNaOption.Value[0:4]))
	}
	if t1 := uint32(binary.BigEndian.Uint32(iaNaOption.Value[4:])); t1 != expectedT1 {
		t.Fatalf("Expected t1 of %d, got %d", expectedT1, t1)
	}
	if t2 := uint32(binary.BigEndian.Uint32(iaNaOption.Value[8:])); t2 != expectedT2 {
		t.Fatalf("Expected t2 of %d, got %d", expectedT2, t2)
	}
	if serializedIaAddrOption := iaNaOption.Value[12:]; string(serializedIaAddrOption) != string(expectedSerializedIaAddrOption) {
		t.Fatalf("Expected serialized ia addr option %v, got %v", expectedSerializedIaAddrOption, serializedIaAddrOption)
	}
}

func TestMakeStatusOption(t *testing.T) {
	expectedMessage := "Boom!"
	expectedStatusCode := uint16(2)
	noAddrOption := MakeStatusOption(expectedStatusCode, expectedMessage)

	if noAddrOption.Id != OptStatusCode {
		t.Fatalf("Expected option id %d, got %d", OptStatusCode, noAddrOption.Id)
	}
	if noAddrOption.Length != uint16(2 + len(expectedMessage)) {
		t.Fatalf("Expected option length of %d, got %d", 2 + len(expectedMessage), noAddrOption.Length)
	}
	if binary.BigEndian.Uint16(noAddrOption.Value[0:2]) != expectedStatusCode {
		t.Fatalf("Expected status code 2, got %d", binary.BigEndian.Uint16(noAddrOption.Value[0:2]))
	}
	if string(noAddrOption.Value[2:]) != expectedMessage {
		t.Fatalf("Expected message %s, got %s", expectedMessage, string(noAddrOption.Value[2:]))
	}
}

func TestUnmarshalFailsIfOROLengthIsOdd(t *testing.T) {
	in := []byte{0, 6, 0, 3, 0, 1, 1}
	if _, err := MakeOptions(in); err == nil {
		t.Fatalf("Parsing options should fail: option request for options has odd length.")
	}
}

func TestMakeDNSServersOption(t *testing.T) {
	expectedAddress1 := net.ParseIP("2001:db8:f00f:cafe::99")
	expectedAddress2 := net.ParseIP("2001:db8:f00f:cafe::9A")
	dnsServersOption := MakeDNSServersOption([]net.IP{expectedAddress1, expectedAddress2})

	if dnsServersOption.Id != OptRecursiveDns {
		t.Fatalf("Expected option id %d, got %d", OptRecursiveDns, dnsServersOption.Id)
	}
	if dnsServersOption.Length != 32 {
		t.Fatalf("Expected length 32 bytes, got %d", dnsServersOption.Length)
	}
	if string(dnsServersOption.Value[0:16]) != string(expectedAddress1) {
		t.Fatalf("Expected dns server address %v, got %v", expectedAddress1, net.IP(dnsServersOption.Value[0:16]))
	}
	if string(dnsServersOption.Value[16:]) != string(expectedAddress2) {
		t.Fatalf("Expected dns server address %v, got %v", expectedAddress2, net.IP(dnsServersOption.Value[16:]))
	}
}
