package dhcp6

import (
	"testing"
	"net"
	"time"
)

func TestReserveAddress(t *testing.T) {
	expectedIP1 := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedIP2 := net.ParseIP("2001:db8:f00f:cafe::2")
	expectedClientID := []byte("Client-id")
	expectedIAID1 := []byte("interface-id-1")
	expectedIAID2 := []byte("interface-id-2")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(expectedIP1, 2, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	ias, _ := pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID1, expectedIAID2})

	if len(ias) != 2 {
		t.Fatalf("Expected 2 identity associations but received %d", len(ias))
	}
	if string(ias[0].IPAddress) !=  string(expectedIP1) && string(ias[0].IPAddress) !=  string(expectedIP2) {
		t.Fatalf("Unexpected ip address: %v", ias[0].IPAddress)
	}
	if string(ias[0].ClientID) !=  string(expectedClientID) {
		t.Fatalf("Expected Client id: %v, but got: %v", expectedClientID, ias[0].ClientID)
	}
	if string(ias[0].InterfaceID) !=  string(expectedIAID1) {
		t.Fatalf("Expected interface id: %v, but got: %v", expectedIAID1, ias[0].InterfaceID)
	}
	if ias[0].CreatedAt != expectedTime {
		t.Fatalf("Expected creation time: %v, but got: %v", expectedTime, ias[0].CreatedAt)
	}

	if string(ias[1].IPAddress) !=  string(expectedIP1) && string(ias[1].IPAddress) !=  string(expectedIP2) {
		t.Fatalf("Unexpected ip address: %v", ias[0].IPAddress)
	}
	if string(ias[1].ClientID) !=  string(expectedClientID) {
		t.Fatalf("Expected Client id: %v, but got: %v", expectedClientID, ias[1].ClientID)
	}
	if string(ias[1].InterfaceID) !=  string(expectedIAID2) {
		t.Fatalf("Expected interface id: %v, but got: %v", expectedIAID2, ias[1].InterfaceID)
	}
	if ias[1].CreatedAt != expectedTime {
		t.Fatalf("Expected creation time: %v, but got: %v", expectedTime, ias[1].CreatedAt)
	}
}

func TestReserveAddressUpdatesAddressPool(t *testing.T) {
	expectedClientID := []byte("Client-id")
	expectedIAID := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID})
	expectedIdx := pool.calculateIAIDHash(expectedClientID, expectedIAID)

	a, exists := pool.identityAssociations[expectedIdx]
	if !exists {
		t.Fatalf("Expected to find identity association at %d but didn't", expectedIdx)
	}
	if string(a.ClientID) != string(expectedClientID) || string(a.InterfaceID) != string(expectedIAID) {
		t.Fatalf("Expected ia association with Client id %x and ia id %x, but got %x %x respectively",
			expectedClientID, expectedIAID, a.ClientID, a.InterfaceID)
	}
}

func TestReserveAddressKeepsTrackOfUsedAddresses(t *testing.T) {
	expectedClientID := []byte("Client-id")
	expectedIAID := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID})

	_, exists := pool.usedIps[0x01]; if !exists {
		t.Fatal("'2001:db8:f00f:cafe::1' should be marked as in use")
	}
}

func TestReserveAddressKeepsTrackOfAssociationExpiration(t *testing.T) {
	expectedClientID := []byte("Client-id")
	expectedIAID := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID})

	expiration := pool.identityAssociationExpirations.Peek().(*associationExpiration)
	if expiration == nil {
		t.Fatal("Expected an identity association expiration, but got nil")
	}
	if expiration.expiresAt != pool.calculateAssociationExpiration(expectedTime) {
		t.Fatalf("Expected association to expire at %v, but got %v",
			pool.calculateAssociationExpiration(expectedTime), expiration.expiresAt)
	}
}

func TestReserveAddressReturnsExistingAssociation(t *testing.T) {
	expectedClientID := []byte("Client-id")
	expectedIAID := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	firstAssociation, _ := pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID})
	secondAssociation, _ := pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID})

	if len(firstAssociation) < 1 {
		t.Fatalf("No associations returned from the first call to ReserveAddresses")
	}
	if len(secondAssociation) < 1 {
		t.Fatalf("No associations returned from the second call to ReserveAddresses")
	}
	if string(firstAssociation[0].IPAddress) != string(secondAssociation[0].IPAddress) {
		t.Fatal("Expected return of the same ip address on both invocations")
	}
}

func TestReleaseAddress(t *testing.T) {
	expectedClientID := []byte("Client-id")
	expectedIAID := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	a, _ := pool.ReserveAddresses(expectedClientID, [][]byte{expectedIAID})

	pool.ReleaseAddresses(expectedClientID, [][]byte{expectedIAID})

	_, exists := pool.identityAssociations[pool.calculateIAIDHash(expectedClientID, expectedIAID)]; if exists {
		t.Fatalf("identity association for %v should've been removed, but is still available", a[0].IPAddress)
	}
}
