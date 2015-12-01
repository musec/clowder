package server

import (
	"net"
	"time"
	"encoding/binary"
)

//IP Addresses status
type IPStat byte

const (
	AVAILABLE	IPStat = 0	//
	RESERVED	IPStat = 1	//	Offered/ MAC Address binding
	ALLOCATED	IPStat = 2	//
)

type Lease struct {
	ip	net.IP
	stat	IPStat
	mac	net.HardwareAddr
	expiry	time.Time
}

type Leases []Lease

func NewLeases(start net.IP, ipRange int) Leases {
	leases := make(Leases,ipRange,ipRange)
	for i :=0; i< iprange; i++{
		ip := AddIP(start,i)
		leases[i]=Lease{ip,0,net.HardwareAddr{},time.Time{}}
	}
	return leases
}

func (l Leases) GetLease(ip net.IP) *Lease {
	for _,l:= range o.leases {
		if l.ip.Equal(ip) {
			return &l

	for i:=0; i<p.ipRange; i++ {
		if IpToUint32(p.leases[i].ip)==IpToUint32(ip) {
			return &p.leases[i]
		}
	}
	return nil
}

func (p IPPool) GetAvailIP() net.IP {
	for i:=0;i<p.ipRange; i++ {
		if p.leases[i].stat== AVAILABLE {
			return p.leases[i].ip
		}
	}
	return nil
}

func (p *IPPool) SetIPStat(ip net.IP, stat IPStat) {
	p.GetLease(ip).stat = stat
}

func (p *IPPool) SetMAC(ip net.IP, mac net.HardwareAddr) {
	p.GetLease(ip).mac = mac
}

func (p *IPPool) SetExpTime(ip net.IP, t time.Time) {
	p.GetLease(ip).expiry = t
}


func AddIP(start net.IP, n int) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, IpToUint32(start)+uint32(n))
	return ip
}

func IpToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

func (p *IPPool) Export() string {
	result:=""
	for l
