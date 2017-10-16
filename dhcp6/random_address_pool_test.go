package dhcp6

import (
	"testing"
	"net"
	"time"
)

func TestReserveAddress(t *testing.T) {
	expectedIp := net.ParseIP("2001:db8:f00f:cafe::1")
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	ia := pool.ReserveAddress(expectedClientId, expectedIaId)

	if ia == nil {
		t.Fatalf("Expected a non-nil identity association")
	}
	if string(ia.ipAddress) !=  string(expectedIp) {
		t.Fatalf("Expected ip: %v, but got: %v", expectedIp, ia.ipAddress)
	}
	if string(ia.clientId) !=  string(expectedClientId) {
		t.Fatalf("Expected client id: %v, but got: %v", expectedClientId, ia.clientId)
	}
	if string(ia.interfaceId) !=  string(expectedIaId) {
		t.Fatalf("Expected interface id: %v, but got: %v", expectedIaId, ia.interfaceId)
	}
	if ia.createdAt != expectedTime {
		t.Fatalf("Expected creation time: %v, but got: %v", expectedTime, ia.createdAt)
	}
	if ia.createdAt != expectedTime {
		t.Fatalf("Expected creation time: %v, but got: %v", expectedTime, ia.createdAt)
	}
	expectedT1 := pool.calculateT1(expectedMaxLifetime); if ia.t1 != expectedT1 {
		t.Fatalf("Expected creation t1: %v, but got: %v", expectedT1, ia.t1)
	}
	expectedT2 := pool.calculateT2(expectedMaxLifetime); if ia.t2 != expectedT2 {
		t.Fatalf("Expected creation t2: %v, but got: %v", expectedT2, ia.t2)
	}
}

func TestReserveAddressUpdatesAddressPool(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddress(expectedClientId, expectedIaId)
	expectedIdx := pool.calculateIaIdHash(expectedClientId, expectedIaId)


	a, exists := pool.identityAssociations[expectedIdx]
	if !exists {
		t.Fatalf("Expected to find identity association at %d but didn't", expectedIdx)
	}
	if string(a.clientId) != string(expectedClientId) || string(a.interfaceId) != string(expectedIaId) {
		t.Fatalf("Expected ia association with client id %x and ia id %x, but got %x %x respectively", expectedClientId, expectedIaId, a.clientId, a.interfaceId)
	}
}

func TestReserveAddressKeepsTrackOfUsedAddresses(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddress(expectedClientId, expectedIaId)

	_, exists := pool.usedIps[0x01]; if !exists {
		t.Fatal("'2001:db8:f00f:cafe::1' should be marked as in use")
	}
}

func TestReserveAddressKeepsTrackOfAssociationExpiration(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	pool.ReserveAddress(expectedClientId, expectedIaId)

	expiration := pool.identityAssociationExpirations.Peek().(*AssociationExpiration)
	if expiration == nil {
		t.Fatal("Expected an identity association expiration, but got nil")
	}
	if expiration.expiresAt != pool.calculateAssociationExpiration(expectedTime, expectedMaxLifetime) {
		t.Fatalf("Expected association to expire at %v, but got %v",
			pool.calculateAssociationExpiration(expectedTime, expectedMaxLifetime), expiration.expiresAt)
	}
}

func TestReserveAddressReturnsExistingAssociation(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	firstAssociation := pool.ReserveAddress(expectedClientId, expectedIaId)
	secondAssociation := pool.ReserveAddress(expectedClientId, expectedIaId)

	if string(firstAssociation.ipAddress) != string(secondAssociation.ipAddress) {
		t.Fatal("Expected return of the same ip address on both invocations")
	}
}

func TestReleaseAddress(t *testing.T) {
	expectedClientId := []byte("client-id")
	expectedIaId := []byte("interface-id")
	expectedTime := time.Now()
	expectedMaxLifetime := uint32(100)

	pool := NewRandomAddressPool(net.ParseIP("2001:db8:f00f:cafe::1"), net.ParseIP("2001:db8:f00f:cafe::1"), expectedMaxLifetime)
	pool.timeNow = func() time.Time { return expectedTime }
	a := pool.ReserveAddress(expectedClientId, expectedIaId)

	pool.ReleaseAddress(expectedClientId, expectedIaId, a.ipAddress)

	_, exists := pool.identityAssociations[pool.calculateIaIdHash(expectedClientId, expectedIaId)]; if exists {
		t.Fatalf("identity association for %v should've been removed, but is still available", a.ipAddress)
	}
}
