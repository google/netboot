package dhcp6

import (
	"net"
	"math/rand"
	"time"
	"math/big"
	"hash/fnv"
	"sync"
	"fmt"
)

type AssociationExpiration struct {
	expiresAt time.Time
	ia *IdentityAssociation
}

type Fifo struct {q []interface{}}

func newFifo() Fifo {
	return Fifo{q: make([]interface{}, 0, 1000)}
}

func (f *Fifo) Push(v interface{}) {
	f.q = append(f.q, v)
}

func (f *Fifo) Shift() interface{} {
	var ret interface{}
	ret, f.q = f.q[0], f.q[1:]
	return ret
}

func (f *Fifo) Size() int {
	return len(f.q)
}

func (f *Fifo) Peek() interface{} {
	if len(f.q) == 0 {
		return nil
	}
	return f.q[0]
}

type RandomAddressPool struct {
	poolStartAddress               *big.Int
	poolSize                 	   uint64
	identityAssociations           map[uint64]*IdentityAssociation
	usedIps                        map[uint64]struct{}
	identityAssociationExpirations Fifo
	validLifetime                  uint32 // in seconds
	timeNow                        func() time.Time
	lock                           sync.Mutex
}

func NewRandomAddressPool(poolStartAddress net.IP, poolSize uint64, validLifetime uint32) *RandomAddressPool {
	ret := &RandomAddressPool{}
	ret.validLifetime = validLifetime
	ret.poolStartAddress = big.NewInt(0)
	ret.poolStartAddress.SetBytes(poolStartAddress)
	ret.poolSize = poolSize
	ret.identityAssociations = make(map[uint64]*IdentityAssociation)
	ret.usedIps = make(map[uint64]struct{})
	ret.identityAssociationExpirations = newFifo()
	ret.timeNow = func() time.Time { return time.Now() }

	ticker := time.NewTicker(time.Second * 10).C
	go func() {
		for {
			<- ticker
			ret.ExpireIdentityAssociations()
		}
	}()

	return ret
}

func (p *RandomAddressPool) ReserveAddresses(clientID []byte, interfaceIDs [][]byte) ([]*IdentityAssociation, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	ret := make([]*IdentityAssociation, 0, len(interfaceIDs))
	rng := rand.New(rand.NewSource(p.timeNow().UnixNano()))

	for _, interfaceID := range (interfaceIDs) {
		clientIDHash := p.calculateIAIDHash(clientID, interfaceID)
		association, exists := p.identityAssociations[clientIDHash];

		if exists {
			ret = append(ret, association)
			continue
		}
		if uint64(len(p.usedIps)) == p.poolSize {
			return ret, fmt.Errorf("No more free ip addresses are currently available in the pool")
		}

		for {
			// we assume that ip addresses adhere to high 64 bits for net and subnet ids, low 64 bits are for host id rule
			hostOffset := randomUint64(rng) % p.poolSize
			newIp := big.NewInt(0).Add(p.poolStartAddress, big.NewInt(0).SetUint64(hostOffset))
			_, exists := p.usedIps[newIp.Uint64()];
			if !exists {
				timeNow := p.timeNow()
				association := &IdentityAssociation{ClientID: clientID,
					InterfaceID: interfaceID,
					IPAddress: newIp.Bytes(),
					CreatedAt: timeNow}
				p.identityAssociations[clientIDHash] = association
				p.usedIps[newIp.Uint64()] = struct{}{}
				p.identityAssociationExpirations.Push(&AssociationExpiration{expiresAt: p.calculateAssociationExpiration(timeNow), ia: association})
				ret = append(ret, association)
				break
			}
		}
	}

	return ret, nil
}

func  (p *RandomAddressPool) ReleaseAddresses(clientID []byte, interfaceIDs [][]byte) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, interfaceId := range(interfaceIDs) {
		association, exists := p.identityAssociations[p.calculateIAIDHash(clientID, interfaceId)]
		if !exists {
			continue
		}
		delete(p.usedIps, big.NewInt(0).SetBytes(association.IPAddress).Uint64())
		delete(p.identityAssociations, p.calculateIAIDHash(clientID, interfaceId))
	}
}

func (p *RandomAddressPool) ExpireIdentityAssociations() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		if p.identityAssociationExpirations.Size() < 1 { break }
		expiration := p.identityAssociationExpirations.Peek().(*AssociationExpiration)
		if p.timeNow().Before(expiration.expiresAt) { break }
		p.identityAssociationExpirations.Shift()
		delete(p.identityAssociations, p.calculateIAIDHash(expiration.ia.ClientID, expiration.ia.InterfaceID))
		delete(p.usedIps, big.NewInt(0).SetBytes(expiration.ia.IPAddress).Uint64())
	}
}

func (p *RandomAddressPool) calculateAssociationExpiration(now time.Time) time.Time {
	return now.Add(time.Duration(p.validLifetime)*time.Second)
}

func (p *RandomAddressPool) calculateIAIDHash(clientID, interfaceID []byte) uint64 {
	h := fnv.New64a()
	h.Write(clientID)
	h.Write(interfaceID)
	return h.Sum64()
}

func randomUint64(rng *rand.Rand) uint64 {
	return uint64(rng.Uint32())<<32 + uint64(rand.Uint32())
}
