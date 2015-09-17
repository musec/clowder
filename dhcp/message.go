package DHCP

import (
	"net"
)

type Packet []byte

//op(1):  Message op code / message type. 1 = BOOTREQUEST, 2 = BOOTREPLY
func (p Packet) OpCode() byte	{ return p[0] }

//htype(1): Hardware address type, see ARP section in "Assigned
//Numbers" RFC; e.g., '1' = 10mb ethernet.
func (p Packet) HType() byte	{ return p[1] }

//hlen(1): Hardware address length (e.g.  '6' for 10mb ethernet).
func (p Packet) HLen() byte		{ return p[2] }

//hops(1): Client sets to zero, optionally used by relay agents
//when booting via a relay agent.
func (p Packet) Hops() byte		{ return p[3] }

//xid(4): Transaction ID, a random number chosen by the
//client, used by the client and server to associate
//messages and responses between a client and a server.
func (p Packet) XId() []byte	{ return p[4:8] }

//secs(2): Filled in by client, seconds elapsed since client
//began address acquisition or renewal process.
func (p Packet) Secs() []byte	{ return p[8:10] }	

//flags(2): Flags
//                                    1 1 1 1 1 1
//                0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
//                +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//                |B|             MBZ             |
//                +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
//                B:  BROADCAST flag
//
//                MBZ:  MUST BE ZERO (reserved for future use)
func (p Packet) Flags() []byte	{ return p[10:12] }

//ciaddr(4): Client IP address; only filled in if client is in
//BOUND, RENEW or REBINDING state and can respond to ARP requests.
func (p Packet) CIAddr() net.IP	{ return net.IP(p[12:16]) }

//yiaddr(4): 'your' (client) IP address.
func (p Packet) YIAddr() net.IP	{ return net.IP(p[16:20]) }

//siaddr(4): IP address of next server to use in bootstrap;
//returned in DHCPOFFER, DHCPACK by server.
func (p Packet) SIAddr() net.IP	{ return net.IP(p[20:24]) }

//giaddr(4): Relay agent IP address, used in booting via a
//relay agent.
func (p Packet) GIAddr() net.IP	{ return net.IP(p[24:28]) }

//chaddr(16): Client hardware address.
func (p Packet) CHAddr() net.HardwareAddr {
	//TODO
	return nil
}

//sname(64) Optional server host name, null terminated string.
func (p Packet) SName() []byte { return p[44:108] }

//file(128): Boot file name, null terminated string; "generic"
//name or null in DHCPDISCOVER, fully qualified
//directory-path name in DHCPOFFER.
func (p Packet) File() []byte { return p[108:236] }

//options(var): Optional parameters field.  See the options
//documents for a list of defined options.
func (p Packet) Options() []byte {
	if len(p) > 236 {
		return p[236:]
	}
	return nil
}

func (p Packet) SetOpCode(opCode byte) { p[0] = byte(opCode) }
func (p Packet) SetCHAddr(hAdd net.HardwareAddr) {
	copy(p[28:44], hAdd)
	p[2] = byte(len(hAdd))
}

func (p Packet) SetHType(hType byte)	{ p[1] = hType }
func (p Packet) SetHops(hops byte)		{ p[3] = hops }
func (p Packet) SetXId(xId []byte)		{ copy(p.XId(), xId) }
func (p Packet) SetSecs(secs []byte)	{ copy(p.Secs(), secs) }
func (p Packet) SetFlags(flags []byte)	{ copy(p.Flags(), flags) }
func (p Packet) SetCIAddr(ip net.IP)	{ copy(p.CIAddr(), ip.To4()) }
func (p Packet) SetYIAddr(ip net.IP)	{ copy(p.YIAddr(), ip.To4()) }
func (p Packet) SetSIAddr(ip net.IP)	{ copy(p.SIAddr(), ip.To4()) }
func (p Packet) SetGIAddr(ip net.IP)	{ copy(p.GIAddr(), ip.To4()) }
func (p Packet) SetSName(sName []byte)	{ copy(p[44:108], sName)}
func (p Packet) SetFile(file []byte)	{
	copy(p[108:236], file)
	if len(file) < 128 {
		p[108+len(file)] = 0
	}
}
func (p Packet) SetOPtions(opt []byte)	{ copy(p[236:len(opt)], opt)
	
}





