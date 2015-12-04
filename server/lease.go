package server

import (
	"net"
	"time"
	"encoding/binary"
	"bytes"
)

//IP Addresses status
type IPStat byte

const (
	AVAILABLE	IPStat = 0	//
	RESERVED	IPStat = 1	//	Offered/ MAC Address binding
	ALLOCATED	IPStat = 2	//
)
var StatMap = map[IPStat]string{
    AVAILABLE: "AVAILABLE",
    RESERVED: "RESERVED",
    ALLOCATED: "ALLOCATED",
}
type Lease struct {
	ip	net.IP
	stat	IPStat
	mac	net.HardwareAddr
	expiry	time.Time
}

type Leases []Lease

func NewLeases(start net.IP, ipRange int) Leases {
	leases := make(Leases,ipRange,ipRange)
	for i :=0; i< ipRange; i++{
		ip := AddIP(start,i)
		leases[i]=Lease{ip,0,net.HardwareAddr{},time.Time{}}
	}
	return leases
}

func (p Leases) GetLease(ip net.IP) *Lease {
	for i:= range p {
		if p[i].ip.Equal(ip) {
			return &p[i]
		}
	}
	return nil
}

func (p Leases) GetLeaseFromMac(mac net.HardwareAddr) *Lease {
	for i:= range p {
		if bytes.Equal(p[i].mac,mac) {
			return &p[i]
		}
	}
	return nil
}

func (p Leases) GetAvailLease() *Lease {
	for i :=range p {
		if p[i].stat== AVAILABLE {
			return &p[i]
		}
	}
	return nil
}

func (p Leases) SetIPStat(ip net.IP, stat IPStat) {
	p.GetLease(ip).stat = stat
}

func (p Leases) SetMac(ip net.IP, mac net.HardwareAddr) {
	p.GetLease(ip).mac = mac
}

func (p Leases) SetExpTime(ip net.IP, t time.Time) {
	p.GetLease(ip).expiry = t
}

func (p Leases) Export() string {
	s:=""
	for _,l:= range p{
		if l.stat!=AVAILABLE{
			s+=l.ip.String()+"\t"+StatMap[l.stat]+"\t"+l.mac.String()+"\t"+l.expiry.String()+"\n"
		}
	}
	if s!="" {
		s = s[:len(s)-1]
	}
	return s
}

func (p Leases) Refresh() {
	now:=time.Now()
	for i:= range p {
		if p[i].stat!=AVAILABLE && now.After(p[i].expiry) {
			p[i].stat=AVAILABLE
			p[i].expiry=time.Time{}
		}
	}
}

func (p Leases) UpdateLease(l Lease) {
	lease:=p.GetLease(l.ip)
	*lease=l
}

func InRange(ip, start net.IP, ipRange int) bool {
	return IpToUint32(ip)>=IpToUint32(start) && IpToUint32(ip)<IpToUint32(start)+uint32(ipRange)
}

func AddIP(start net.IP, n int) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, IpToUint32(start)+uint32(n))
	return ip
}

func IpToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

