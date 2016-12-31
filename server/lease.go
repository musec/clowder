/*
 * Copyright 2015 Nhac Nguyen
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"
)

//IP Addresses status
type IPStat byte

const (
	AVAILABLE    IPStat = 0 //
	RESERVED     IPStat = 1 //	Offered/ MAC Address binding
	ALLOCATED    IPStat = 2 //
	NOTAVAILABLE IPStat = 3 //	Conflicted IP
)

var StatMap = map[IPStat]string{
	AVAILABLE: "AVAILABLE",
	RESERVED:  "RESERVED",
	ALLOCATED: "ALLOCATED",
}

type Lease struct {
	Ip     net.IP
	Stat   IPStat
	Mac    net.HardwareAddr
	Expiry time.Time
}

type Leases []Lease

//NewLeases creates a new Leases
func NewLeases(start net.IP, ipRange int) Leases {
	leases := make(Leases, ipRange, ipRange)
	for i := 0; i < ipRange; i++ {
		ip := AddIP(start, i)
		//fmt.Println(start)
		leases[i] = Lease{ip, 0, net.HardwareAddr{}, time.Time{}}
	}
	return leases
}

//GetLeaseFromMAC returns a pointer to lease having a specific IP address
func (p Leases) GetLease(ip net.IP) *Lease {
	for i := range p {
		if p[i].Ip.Equal(ip) {
			return &p[i]
		}
	}
	return nil
}

//GetLeaseFromMAC returns a pointer to lease having a specific MAC address
func (p Leases) GetLeaseFromMac(mac net.HardwareAddr) *Lease {
	for i := range p {
		if bytes.Equal(p[i].Mac, mac) {
			return &p[i]
		}
	}
	return nil
}

//GetAvailLease returns a pointer to lease which its status is AVAILABLE
func (p Leases) GetAvailLease() *Lease {
	for i := range p {
		if p[i].Stat == AVAILABLE {
			return &p[i]
		}
	}
	return nil
}

func (p Leases) SetIPStat(ip net.IP, stat IPStat) {
	p.GetLease(ip).Stat = stat
}

func (p Leases) SetMac(ip net.IP, mac net.HardwareAddr) {
	p.GetLease(ip).Mac = mac
}

func (p Leases) SetExpTime(ip net.IP, t time.Time) {
	p.GetLease(ip).Expiry = t
}

//Exports generates leases information into a string
func (p Leases) String() string {
	s := ""
	for _, l := range p {
		if l.Stat != AVAILABLE {
			s += l.Ip.String() + "\t" + StatMap[l.Stat] + "\t" + l.Mac.String() + "\t" + l.Expiry.String() + "\n"
		}
	}
	if s != "" {
		s = s[:len(s)-1]
	}
	return s
}

//Refresh checks leases' expiry time. If a lease is expiried leases, its status will be set to AVAILABLE
//and its MAC address will be removed from the record.
//UPDATE: Refresh only check ALLOCATED leases, set their status to RESERVED if they are expiried
func (p Leases) Refresh() {
	now := time.Now()
	for i := range p {
		if p[i].Stat == ALLOCATED && now.After(p[i].Expiry) {
			p[i].Stat = RESERVED
			p[i].Expiry = time.Time{}
			//p[i].Mac=net.HardwareAddr{}
		}
	}
}

func (p Leases) UpdateLease(l Lease) {
	lease := p.GetLease(l.Ip)
	if lease != nil {
		*lease = l
	}
}

func InRange(ip, start net.IP, ipRange int) bool {
	return IpToUint32(ip) >= IpToUint32(start) && IpToUint32(ip) < IpToUint32(start)+uint32(ipRange)
}

func AddIP(start net.IP, n int) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, IpToUint32(start)+uint32(n))
	return ip
}

func IpToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}
