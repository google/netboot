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
	var to_ret interface{}
	to_ret, f.q = f.q[0], f.q[1:]
	return to_ret
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
	to_ret := &RandomAddressPool{}
	to_ret.validLifetime = validLifetime
	to_ret.poolStartAddress = big.NewInt(0)
	to_ret.poolStartAddress.SetBytes(poolStartAddress)
	to_ret.poolSize = poolSize
	to_ret.identityAssociations = make(map[uint64]*IdentityAssociation)
	to_ret.usedIps = make(map[uint64]struct{})
	to_ret.identityAssociationExpirations = newFifo()
	to_ret.timeNow = func() time.Time { return time.Now() }

	ticker := time.NewTicker(time.Second * 10).C
	go func() {
		for {
			<- ticker
			to_ret.ExpireIdentityAssociations()
		}
	}()

	return to_ret
}

func (p *RandomAddressPool) ReserveAddresses(clientId []byte, interfaceIds [][]byte) ([]*IdentityAssociation, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	ret := make([]*IdentityAssociation, 0, len(interfaceIds))
	rng := rand.New(rand.NewSource(p.timeNow().UnixNano()))

	for _, interfaceId := range (interfaceIds) {
		clientIdHash := p.calculateIaIdHash(clientId, interfaceId)
		association, exists := p.identityAssociations[clientIdHash];

		if exists {
			ret = append(ret, association)
			continue
		}
		if uint64(len(p.usedIps)) == p.poolSize {
			return ret, fmt.Errorf("No more free ip addresses are currently available in the pool")
		}

		for {
			// we assume that ip addresses adhere to high 64 bits for net and subnet ids, low 64 bits are for host id rule
			hostOffset := rng.Uint64() % p.poolSize
			newIp := big.NewInt(0).Add(p.poolStartAddress, big.NewInt(0).SetUint64(hostOffset))
			_, exists := p.usedIps[newIp.Uint64()];
			if !exists {
				timeNow := p.timeNow()
				association := &IdentityAssociation{ClientId: clientId,
					InterfaceId: interfaceId,
					IpAddress: newIp.Bytes(),
					CreatedAt: timeNow}
				p.identityAssociations[clientIdHash] = association
				p.usedIps[newIp.Uint64()] = struct{}{}
				p.identityAssociationExpirations.Push(&AssociationExpiration{expiresAt: p.calculateAssociationExpiration(timeNow), ia: association})
				ret = append(ret, association)
				break
			}
		}
	}

	return ret, nil
}

func  (p *RandomAddressPool) ReleaseAddresses(clientId []byte, interfaceIds [][]byte) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, interfaceId := range(interfaceIds) {
		association, exists := p.identityAssociations[p.calculateIaIdHash(clientId, interfaceId)]
		if !exists {
			continue
		}
		delete(p.usedIps, big.NewInt(0).SetBytes(association.IpAddress).Uint64())
		delete(p.identityAssociations, p.calculateIaIdHash(clientId, interfaceId))
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
		delete(p.identityAssociations, p.calculateIaIdHash(expiration.ia.ClientId, expiration.ia.InterfaceId))
		delete(p.usedIps, big.NewInt(0).SetBytes(expiration.ia.IpAddress).Uint64())
	}
}

func (p *RandomAddressPool) calculateAssociationExpiration(now time.Time) time.Time {
	return now.Add(time.Duration(p.validLifetime)*time.Second)
}

func (p *RandomAddressPool) calculateIaIdHash(clientId, interfaceId []byte) uint64 {
	h := fnv.New64a()
	h.Write(clientId)
	h.Write(interfaceId)
	return h.Sum64()
}
