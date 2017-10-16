package dhcp6

import (
	"net"
	"math/rand"
	"time"
	"math/big"
	"hash/fnv"
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
	poolEndAddress                 *big.Int
	identityAssociations           map[uint64]*IdentityAssociation
	usedIps                        map[uint64]struct{}
	identityAssociationExpirations Fifo
	preferredLifetime              uint32 // in seconds
	timeNow                        func() time.Time
	lock                           chan int
}

func NewRandomAddressPool(poolStartAddress, poolEndAddress net.IP, preferredLifetime uint32) *RandomAddressPool {
	to_ret := &RandomAddressPool{}
	to_ret.preferredLifetime = preferredLifetime
	to_ret.poolStartAddress = big.NewInt(0)
	to_ret.poolStartAddress.SetBytes(poolStartAddress)
	to_ret.poolEndAddress = big.NewInt(0)
	to_ret.poolEndAddress.SetBytes(poolEndAddress)
	to_ret.identityAssociations = make(map[uint64]*IdentityAssociation)
	to_ret.usedIps = make(map[uint64]struct{})
	to_ret.identityAssociationExpirations = newFifo()
	to_ret.timeNow = func() time.Time { return time.Now() }
	to_ret.lock = make(chan int, 1)
	to_ret.lock <- 1

	ticker := time.NewTicker(time.Second * 10).C
	go func() {
		for {
			<- ticker
			to_ret.ExpireIdentityAssociations()
		}
	}()

	return to_ret
}

func (p *RandomAddressPool) ReserveAddress(clientId, interfaceId []byte) *IdentityAssociation {
	<-p.lock
	clientIdHash := p.calculateIaIdHash(clientId, interfaceId)
	association, exists := p.identityAssociations[clientIdHash]; if exists {
		p.lock <- 1
		return association
	}

	for {
		rng := rand.New(rand.NewSource(p.timeNow().UnixNano()))
		// we assume that ip addresses adhere to high 64 bits for net and subnet ids, low 64 bits are for host id rule
		hostOffset := rng.Uint64()%(p.poolEndAddress.Uint64() - p.poolStartAddress.Uint64() + 1)
		newIp := big.NewInt(0).Add(p.poolStartAddress, big.NewInt(0).SetUint64(hostOffset))
		_, exists := p.usedIps[newIp.Uint64()]; if !exists {
			timeNow := p.timeNow()
			to_ret := &IdentityAssociation{clientId: clientId,
				interfaceId: interfaceId,
				ipAddress: newIp.Bytes(),
				createdAt: timeNow,
				t1: p.calculateT1(p.preferredLifetime),
				t2: p.calculateT2(p.preferredLifetime) }
			p.identityAssociations[clientIdHash] = to_ret
			p.usedIps[newIp.Uint64()] = struct{}{}
			p.identityAssociationExpirations.Push(&AssociationExpiration{expiresAt: p.calculateAssociationExpiration(timeNow, p.preferredLifetime), ia: to_ret})
			p.lock <- 1
			return to_ret
		}
	}
	p.lock <- 1
	return nil
}

func  (p *RandomAddressPool) ReleaseAddress(clientId, interfaceId []byte, addr net.IP) {
	<-p.lock
	delete(p.identityAssociations, p.calculateIaIdHash(clientId, interfaceId))
	delete(p.usedIps, big.NewInt(0).SetBytes(addr).Uint64())
	p.lock <- 1
}

func (p *RandomAddressPool) ExpireIdentityAssociations() {
	<-p.lock
	for {
		if p.identityAssociationExpirations.Size() < 1 { break }
		expiration := p.identityAssociationExpirations.Peek().(*AssociationExpiration)
		if p.timeNow().Before(expiration.expiresAt) { break }
		p.identityAssociationExpirations.Shift()
		delete(p.identityAssociations, p.calculateIaIdHash(expiration.ia.clientId, expiration.ia.interfaceId))
		delete(p.usedIps, big.NewInt(0).SetBytes(expiration.ia.ipAddress).Uint64())
	}
	p.lock <- 1
}

func (p *RandomAddressPool) calculateT1(preferredLifetime uint32) uint32 {
 return preferredLifetime / 2
}

func (p *RandomAddressPool) calculateT2(preferredLifetime uint32) uint32 {
 return (preferredLifetime * 4)/5
}

func (p *RandomAddressPool) calculateAssociationExpiration(now time.Time, preferredLifetime uint32) time.Time {
	return now.Add(time.Duration(p.preferredLifetime)*time.Second)
}

func (p *RandomAddressPool) calculateIaIdHash(clientId, interfaceId []byte) uint64 {
	h := fnv.New64a()
	h.Write(clientId)
	h.Write(interfaceId)
	return h.Sum64()
}
