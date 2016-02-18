package pxedhcp

import (
	"net"
)

type Packet []byte
type Options map[byte][]byte //map Option's Code -> Option's Value

//
// Create a DHCP REQUEST packet.
//
func NewRequestPacket(xid []byte, broadcast bool, clientIP net.IP, clientMac net.HardwareAddr) Packet {
	p := make(Packet, 241)
	p.setCookie()

	p.SetRequest()
	p.SetHWAddrType(ETHERNET)
	p.SetHardwareAddress(clientMac)
	p.SetTransactionID(xid)
	if clientIP != nil {
		p.SetCurrentClientIP(clientIP)
	}
	p.SetBroadcast(broadcast)
	p[240] = byte(END)
	return p
}

//
// Create a reply packet with some headers copied from a request packet.
//
// See Table 3.
//
func NewReplyPacket(req Packet) Packet {
	p := make(Packet, 241)
	p.setCookie()

	p.SetReply()
	p.SetHWAddrType(ETHERNET)
	p.SetTransactionID(req.TransactionID())
	p.SetFlags(req.Flags())
	p.SetHardwareAddress(req.HardwareAddress())
	p.SetRelayIP(req.RelayIP())
	p[240] = byte(END)
	return p
}

// Is this a DHCP REQUEST packet?
func (p Packet) Request() bool { return p.op()[0] == BOOTREQUEST }
func (p Packet) SetRequest()   { p.op()[0] = BOOTREQUEST }

// Is this a DHCP REPLY packet?
func (p Packet) Reply() bool { return p.op()[0] == BOOTREPLY }
func (p Packet) SetReply()   { p.op()[0] = BOOTREPLY }

// Client-originated transaction ID.
func (p Packet) TransactionID() []byte      { return p.XId() }
func (p Packet) SetTransactionID(id []byte) { copy(p.XId(), id) }

// Client hardware (MAC) address.
func (p Packet) HardwareAddress() net.HardwareAddr {
	len := p.hlen()[0]
	if len > 16 {
		len = 16
	}
	return net.HardwareAddr(p[28 : 28+len])
}

func (p Packet) SetHardwareAddress(addr net.HardwareAddr) {
	copy(p.chaddr(), addr)
	p.hlen()[0] = byte(len(addr))
}

// Type of hardware address (only ETHERNET currently supported).
func (p Packet) HWAddrType() byte         { return p.htype()[0] }
func (p Packet) SetHWAddrType(hType byte) { p.htype()[0] = hType }

// Pre-existing client IP address (e.g., during renewal).
func (p Packet) CurrentClientIP() net.IP      { return net.IP(p.ciaddr()) }
func (p Packet) SetCurrentClientIP(ip net.IP) { copy(p.ciaddr(), ip) }

// IP address of relay server (if present).
func (p Packet) RelayIP() net.IP      { return net.IP(p.giaddr()) }
func (p Packet) SetRelayIP(ip net.IP) { copy(p.giaddr(), ip) }

// Relay hops: should be set to 0 by client.
func (p Packet) Hops() byte        { return p.hops()[0] }
func (p Packet) SetHops(hops byte) { p.hops()[0] = hops }

// Time since transaction began.
func (p Packet) SecondsElapsed() []byte     { return p.secs() }
func (p Packet) SetSecsElapsed(secs []byte) { copy(p.secs(), secs) }

// DHCP flags (which should currently all be zero, except for Broadcast).
func (p Packet) Flags() []byte         { return p.flags() }
func (p Packet) SetFlags(flags []byte) { copy(p.flags(), flags) }

// Broadcast flag: "please broadcast DHCPOFFER responses to the whole network"
func (p Packet) Broadcast() bool { return p.flags()[0]&0x80 != 0 }
func (p Packet) SetBroadcast(broadcast bool) {
	if broadcast {
		p.flags()[0] |= 0x80
	} else {
		p.flags()[0] &= 0x80
	}
}

// DHCP-provided client IP address.
func (p Packet) ClientIP(ip net.IP) net.IP { return net.IP(p.yiaddr()) }
func (p Packet) SetClientIP(ip net.IP)     { copy(p.yiaddr(), ip) }

// DHCP-provided server to contact next in the boot process.
func (p Packet) NextServer(ip net.IP) net.IP { return net.IP(p.siaddr()) }
func (p Packet) SetNextServer(ip net.IP)     { copy(p.siaddr(), ip) }

// Optional DHCP-provided server hostname.
func (p Packet) ServerName(name string) string { return string(p.sname()) }
func (p Packet) SetServerName(name string)     { copy(p.sname(), []byte(name)) }

// Path to PXE boot file.
func (p Packet) BootFile(file string) string { return string(p.file()) }
func (p Packet) SetBootFile(file string)     { copy(p.file(), []byte(file)) }

// Convert 'options' field into a byte->byte map.
func (p Packet) Options() Options {
	opts := make(Options, 10)
	op := p.options()[4:]
	for len(op) >= 0 && op[0] != END {
		if op[0] == PAD {
			op = op[1:]
			continue
		}
		opLen := int(op[1])
		if len(op) < 2+opLen {
			break
		}
		opts[op[0]] = op[2 : 2+opLen]
		op = op[2+opLen:]
	}
	return opts
}

// Add a DHCP option to a packet
func (p *Packet) AddOption(OptCode byte, OptValue []byte) {
	opt := append([]byte{OptCode, byte(len(OptValue))}, OptValue...)
	opt = append(opt, END)
	*p = append((*p)[:len(*p)-1], opt...)
}

// Set all DHCP options.
func (p Packet) SetOptions(opt []byte) { copy(p[236:len(opt)], opt) }

// Zero-pad packet to a defined length.
func (p *Packet) Padding(size int) {
	for len(*p) < size {
		*p = append(*p, PAD)
	}
}
