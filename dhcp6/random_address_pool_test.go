package dhcp6

import (
	"testing"
	"net"
	"time"
)

func TestReserveAddress(t *testing.T) {
	expectedIp1 := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedIp2 := net.ParseIP("2001:db8:f00f:cafe::2")
	expectedClientId := []byte("client-id")
	expectedIaId1 := []byte("interface-id-1")
	expectedIaId2 := []byte("interface-id-2")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(expectedIp1, 2, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	ias, _ := pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId1, expectedIaId2})

	if len(ias) != 2 {
		t.Fatalf("Expected 2 identity associations but received %d", len(ias))
	}
	if string(ias[0].IpAddress) !=  string(expectedIp1) && string(ias[0].IpAddress) !=  string(expectedIp2) {
		t.Fatalf("Unexpected ip address: %v", ias[0].IpAddress)
	}
	if string(ias[0].ClientId) !=  string(expectedClientId) {
		t.Fatalf("Expected client id: %v, but got: %v", expectedClientId, ias[0].ClientId)
	}
	if string(ias[0].InterfaceId) !=  string(expectedIaId1) {
		t.Fatalf("Expected interface id: %v, but got: %v", expectedIaId1, ias[0].InterfaceId)
	}
	if ias[0].CreatedAt != expectedTime {
		t.Fatalf("Expected creation time: %v, but got: %v", expectedTime, ias[0].CreatedAt)
	}

	if string(ias[1].IpAddress) !=  string(expectedIp1) && string(ias[1].IpAddress) !=  string(expectedIp2) {
		t.Fatalf("Unexpected ip address: %v", ias[0].IpAddress)
	}
	if string(ias[1].ClientId) !=  string(expectedClientId) {
		t.Fatalf("Expected client id: %v, but got: %v", expectedClientId, ias[1].ClientId)
	}
	if string(ias[1].InterfaceId) !=  string(expectedIaId2) {
		t.Fatalf("Expected interface id: %v, but got: %v", expectedIaId2, ias[1].InterfaceId)
	}
	if ias[1].CreatedAt != expectedTime {
		t.Fatalf("Expected creation time: %v, but got: %v", expectedTime, ias[1].CreatedAt)
	}
}

func TestReserveAddressUpdatesAddressPool(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId})
	expectedIdx := pool.calculateIaIdHash(expectedClientId, expectedIaId)

	a, exists := pool.identityAssociations[expectedIdx]
	if !exists {
		t.Fatalf("Expected to find identity association at %d but didn't", expectedIdx)
	}
	if string(a.ClientId) != string(expectedClientId) || string(a.InterfaceId) != string(expectedIaId) {
		t.Fatalf("Expected ia association with client id %x and ia id %x, but got %x %x respectively",
			expectedClientId, expectedIaId, a.ClientId, a.InterfaceId)
	}
}

func TestReserveAddressKeepsTrackOfUsedAddresses(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId})

	_, exists := pool.usedIps[0x01]; if !exists {
		t.Fatal("'2001:db8:f00f:cafe::1' should be marked as in use")
	}
}

func TestReserveAddressKeepsTrackOfAssociationExpiration(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId})

	expiration := pool.identityAssociationExpirations.Peek().(*AssociationExpiration)
	if expiration == nil {
		t.Fatal("Expected an identity association expiration, but got nil")
	}
	if expiration.expiresAt != pool.calculateAssociationExpiration(expectedTime) {
		t.Fatalf("Expected association to expire at %v, but got %v",
			pool.calculateAssociationExpiration(expectedTime), expiration.expiresAt)
	}
}

func TestReserveAddressReturnsExistingAssociation(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	firstAssociation, _ := pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId})
	secondAssociation, _ := pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId})

	if len(firstAssociation) < 1 {
		t.Fatalf("No associations returned from the first call to ReserveAddresses")
	}
	if len(secondAssociation) < 1 {
		t.Fatalf("No associations returned from the second call to ReserveAddresses")
	}
	if string(firstAssociation[0].IpAddress) != string(secondAssociation[0].IpAddress) {
		t.Fatal("Expected return of the same ip address on both invocations")
	}
}

func TestReleaseAddress(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), 1, expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	a, _ := pool.ReserveAddresses(expectedClientId, [][]byte{expectedIaId})

	pool.ReleaseAddresses(expectedClientId, [][]byte{expectedIaId})

	_, exists := pool.identityAssociations[pool.calculateIaIdHash(expectedClientId, expectedIaId)]; if exists {
		t.Fatalf("identity association for %v should've been removed, but is still available", a[0].IpAddress)
	}
}
