package pxedhcp

import (
	"net"
)

type Packet []byte
type Options map[byte][]byte  //map Option's Code -> Option's Value

//***************************************************************************************************
//**************************************      ACCESSORS     *****************************************
//***************************************************************************************************

func (p Packet) Op()		[]byte	{ return p[0:1] }
func (p Packet) HType()		[]byte	{ return p[1:2] }
func (p Packet) HLen()		[]byte	{ return p[2:3] }
func (p Packet) Hops()		[]byte	{ return p[3:4] }
func (p Packet) XId()		[]byte	{ return p[4:8] }
func (p Packet) Secs()		[]byte	{ return p[8:10] }
func (p Packet) Flags()		[]byte	{ return p[10:12] }
func (p Packet) CIAddr()	[]byte	{ return p[12:16] }
func (p Packet) YIAddr()	[]byte	{ return p[16:20] }
func (p Packet) SIAddr()	[]byte	{ return p[20:24] }
func (p Packet) GIAddr()	[]byte	{ return p[24:28] }
func (p Packet) CHAddr()	[]byte	{ return p[28:44] }
func (p Packet) SName()		[]byte	{ return p[44:108] }
func (p Packet) File()		[]byte	{ return p[108:236] }
func (p Packet) Cookie()	[]byte	{ return p[236:240] }
func (p Packet) Options()	[]byte	{
	if len(p) > 236 {
		return p[236:]
	}
	return nil
}

//***************************************************************************************************
//***********************************     PUBLIC ACCESSORS     *************************************
//***************************************************************************************************

//SETTER
func (p Packet) SetRequest()	{ p.Op()[0] = BOOTREQUEST }
func (p Packet) SetReply()	{ p.Op()[0] = BOOTREPLY }
func (p Packet) SetClientHardwareAddr(hAddr net.HardwareAddr) {
	copy(p.CHAddr(), hAddr)
	p.HLen()[0] = byte(len(hAddr))
}
func (p Packet) SetHWAddrType(hType byte)	{ p.HType()[0] = hType }
func (p Packet) SetHops(hops byte)	{ p.Hops()[0] = hops }
func (p Packet) SetTransID(id []byte)	{ copy(p.XId(), id) }
func (p Packet) SetSecsElapsed(secs []byte)	{ copy(p.Secs(), secs) }
func (p Packet) SetFlags(flags []byte)	{ copy(p.Flags(), flags) }
func (p Packet) SetCIAddr(ip net.IP)	{ copy(p.CIAddr(), ip) }
func (p Packet) SetYIAddr(ip net.IP)	{ copy(p.YIAddr(), ip) }
func (p Packet) SetSIAddr(ip net.IP)	{ copy(p.SIAddr(), ip) }
func (p Packet) SetGIAddr(ip net.IP)	{ copy(p.GIAddr(), ip) }
func (p Packet) SetServerName(sName string)	{ copy(p.SName(), []byte(sName))}
func (p Packet) SetBootFile(file string)	{ copy(p.File(), []byte(file))}
func (p Packet) SetMagicCookie()	{ copy(p.Cookie(), []byte{99, 130, 83, 99}) }
func (p Packet) SetBroadcast(broadcast bool) {
	if broadcast {
		p.Flags()[0] = 128
	} else {
		p.Flags()[0] = 0
	}
}
func (p Packet) SetOPtions(opt []byte)	{ copy(p[236:len(opt)], opt)}

//GETTER
func (p Packet)	GetXID() 		[]byte	{ return p.XId() }
func (p Packet)	GetHWAddrType()	byte	{ return p.HType()[0] }
func (p Packet) GetHardwareAddr() 	net.HardwareAddr {
	hLen := p.HLen()[0]
	if hLen > 16 {
		hLen = 16
	}
	return net.HardwareAddr(p[28 : 28+hLen])
}
func (p Packet)	GetHops()		byte	{ return p.Hops()[0] }
func (p Packet) GetTransID()		[]byte	{ return p.XId() }
func (p Packet) GetSecsElapsed()	[]byte	{ return p.Secs() }
func (p Packet) GetFlags()		[]byte	{ return p.Flags() }
func (p Packet) GetGIAddr()		net.IP	{ return net.IP(p.GIAddr()) }
func (p Packet) GetCIAddr()		net.IP	{ return net.IP(p.CIAddr()) }


// Creates a request packet
func NewRequestPacket(xid []byte, broadcast bool, ciaddr net.IP,chaddr net.HardwareAddr) Packet {
	p := make(Packet, 241)
	p.SetRequest()
	p.SetHWAddrType(ETHERNET)
	p.SetClientHardwareAddr(chaddr)
	p.SetTransID(xid)
	if ciaddr != nil {
		p.SetCIAddr(ciaddr)
	}
	p.SetBroadcast(broadcast)
	p.SetMagicCookie()
	p[240] = byte(END)
	return p
}


//NewReplyPacket creates a reply packet with header's information from request packet
//The header's fileds are xid, flags, giaddr, chaddr. (See RFC 2131, table 3)
func NewReplyPacket(req Packet) Packet {
	p := make(Packet, 241)
	p.SetReply()
	p.SetHWAddrType(ETHERNET)
	p.SetTransID(req.GetTransID())
	p.SetFlags(req.GetFlags())
	p.SetClientHardwareAddr(req.GetHardwareAddr())
	p.SetGIAddr(req.GetGIAddr())
	p.SetMagicCookie()
	p[240] = byte(END)
	return p
}

//ParseOptions read option filed of a packet, parses it into a map[option code]->value:106 
func (p Packet) ParseOptions() Options {
	opts:=make(Options,10)
	op := p.Options()[4:]
	for len(op) >= 0  && op[0] != END {
		if op[0]== PAD {
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
func (p Packet) AddOption(OptCode byte, OptValue []byte) {
	opt:=append([]byte{OptCode,byte(len(OptValue))},OptValue...)
	opt=append(opt,END)
	p = append((p)[:len(p)-1],opt...)
}

// Padding packet to a size
func (p Packet) Padding(size int) {
	for len(p) < size {
		p = append(p, PAD)
	}
}

//
func (p Packet) isBroadcast() bool { return p.Flags()[0] > 127 }

