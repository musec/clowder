package pxedhcp

//
// Raw data access, using RFC 2131 field names:
//

// Message op code / message type.
func (p Packet) op() []byte { return p[0:1] }

// Hardware address type, see ARP section in
// "Assigned Numbers" RFC; e.g., '1' = 10mb ethernet.
func (p Packet) htype() []byte { return p[1:2] }

// Hardware address length (e.g., '6' for 10mb ethernet).
func (p Packet) hlen() []byte { return p[2:3] }

// Client sets to zero, optionally used by relay agents
// when booting via a relay agent.
func (p Packet) hops() []byte { return p[3:4] }

// Transaction ID, a random number chosen by the client, used by the client
// and server to associate messages and responses between a client and a server.
func (p Packet) XId() []byte { return p[4:8] }

// Filled in by client, seconds elapsed since client
// began address acquisition or renewal process.
func (p Packet) secs() []byte { return p[8:10] }

// Flags: Broadcast flag and 15 zeros (see Figure 2).
func (p Packet) flags() []byte { return p[10:12] }

// Client IP address; only filled in if client is in BOUND, RENEW or REBINDING
// state and can respond to ARP requests.
func (p Packet) ciaddr() []byte { return p[12:16] }

// 'your' (client) IP address.
func (p Packet) yiaddr() []byte { return p[16:20] }

// IP address of next server to use in bootstrap;
// returned in DHCPOFFER, DHCPACK by server.
func (p Packet) siaddr() []byte { return p[20:24] }

// Relay agent IP address, used in booting via a relay agent.
func (p Packet) giaddr() []byte { return p[24:28] }

// Client hardware address.
func (p Packet) chaddr() []byte { return p[28:44] }

// Optional server host name, null terminated string.
func (p Packet) sname() []byte { return p[44:108] }

// Boot file name, null terminated string; "generic" name or null in
// DHCPDISCOVER, fully qualified directory-path name in DHCPOFFER.
func (p Packet) file() []byte { return p[108:236] }

// The first four octets of the 'options' field of the DHCP message
// contain the (decimal) values 99, 130, 83 and 99, respectively (this
// is the same magic cookie as is defined in RFC 1497).
func (p Packet) cookie() []byte { return p.options()[0:4] }
func (p Packet) setCookie()     { copy(p.cookie(), []byte{99, 130, 83, 99}) }

func (p Packet) options() []byte {
	if len(p) > 236 {
		return p[236:]
	}
	return nil
}
