package server

import (
	"encoding/binary"
	"github.com/musec/go-dhcp"
	"net"
	"time"
)

func (s *Server) DHCPResponder(p dhcp.Packet) dhcp.Packet {

	//Get lock
	<-s.TablesAccess
	defer func() { s.TablesAccess <- true }()
	options := p.Options()

	//Get DHCP Message Type
	val, ok := options[dhcp.OptDHCPMsgType]
	if !ok || len(val) != 1 {
		return nil
	}
	msgType := val[0]
	if msgType < dhcp.DISCOVER || msgType > dhcp.INFORM {
		return nil
	}

	//Is a PXE DHCP Packet
	pxeRequest := options[dhcp.OptClassId] != nil && options[dhcp.OptClientSystemArchitecture] != nil && options[dhcp.OptClientNetworkDeviceInterface] != nil && options[dhcp.OptUUIDGUID] != nil
	var pool Leases
	var uuid UUID
	var pxe *PxeRecord
	if pxeRequest {
		if len(options[dhcp.OptUUIDGUID]) == 17 {
			uuid = options[dhcp.OptUUIDGUID][1:]
		}
		pool = s.MachineLeases
		pxe = s.Pxe.GetRecord(uuid)
	} else {
		pool = s.DeviceLeases
	}
	pool.Refresh()

	//Lease duration
	duration := make([]byte, 4)
	binary.BigEndian.PutUint32(duration, uint32(s.LeaseDuration/time.Second))

	//Get MAC address of packet and looking for lease having that address
	mac := p.HardwareAddress()
	lease := pool.GetLeaseFromMac(mac)

	switch msgType {
	case dhcp.DISCOVER:
		s.Log("Get DISCOVER message from " + mac.String())
		if lease == nil { //no record of this MAC address
			s.NewHardware[mac.String()] = uuid
			return nil
		}

		response := dhcp.NewReplyPacket(p)
		//set packet header
		response.SetClientIP(lease.Ip)
		response.SetServerName(s.ServerName)
		//set packet options
		//MUST
		response.AddOption(dhcp.OptDHCPMsgType, []byte{dhcp.OFFER}) //Message Type
		response.AddOption(dhcp.OptAddressTime, duration)           //Lease time
		response.AddOption(dhcp.OptDHCPServerId, []byte(s.Ip))      //Server identifier
		//MAY
		response.AddOption(dhcp.OptSubnetMask, []byte(s.Mask))
		response.AddOption(dhcp.OptRouter, []byte(s.Router))
		response.AddOption(dhcp.OptDomainServer, []byte(s.DNS))
		response.AddOption(dhcp.OptDomainName, []byte(s.DomainName))

		if pxeRequest {
			if pxe != nil {
				response.SetBootFile(pxe.BootFile)
				response.AddOption(dhcp.OptRootPath, []byte(pxe.RootPath))
			} else {
				s.Warn("Receive PXE DHCP DISCOVER packet from a known MAC address: " + mac.String() + ". But there isn't PXE information to response.")
				return nil
			}
		}

		//update leases. UPDATE: dont have to update since DHCP server only response to packet from binded MAC address
		//if lease.stat==AVAILABLE {
		//	lease.stat=RESERVED
		//	lease.expiry=time.Now().Add(time.Minute*5)
		//}

		return response

	case dhcp.REQUEST:
		s.Log("Get REQUEST message from " + mac.String())
		//Is the packet for this server
		if serverId, ok := options[dhcp.OptDHCPServerId]; ok && !net.IP(serverId).Equal(s.Ip) {
			return nil
		}

		if lease == nil { //no record of this MAC address
			//s.NewMachines=append(s.NewMachines,Machines{mac,time.Now(),pxeRequest})
			s.Warn("Receive a DHCP REQUEST packet from an unknown MAC address: " + mac.String() + ".")
			return nil
		}

		requestIP := net.IP(options[dhcp.OptAddressRequest])
		if requestIP == nil { //the request packet is for extending its lease
			requestIP = p.CurrentClientIP()
		}

		response := dhcp.NewReplyPacket(p)

		//If the request IP is different the offer IP
		if requestIP.Equal(net.IPv4zero) || !requestIP.Equal(lease.Ip) {
			s.Log("Request IP " + requestIP.String() + " is different to offer IP " + lease.Ip.String())
			response.AddOption(dhcp.OptDHCPMsgType, []byte{dhcp.NAK})
			return response
		}

		if pxeRequest && pxe == nil {
			s.Warn("Receive PXE DHCP REQUEST packet from a known MAC address: " + mac.String() + ". But there isn't PXE information to response.")
			response.AddOption(dhcp.OptDHCPMsgType, []byte{dhcp.NAK})
			return response
		}

		//set packet header
		response.SetClientIP(lease.Ip)
		response.SetServerName(s.ServerName)
		//set packet options
		//MUST
		response.AddOption(dhcp.OptDHCPMsgType, []byte{dhcp.ACK}) //Message Type
		response.AddOption(dhcp.OptAddressTime, duration)         //Lease time
		response.AddOption(dhcp.OptDHCPServerId, []byte(s.Ip))    //Server identifier
		//MAY
		response.AddOption(dhcp.OptSubnetMask, []byte(s.Mask))
		response.AddOption(dhcp.OptRouter, []byte(s.Router))
		response.AddOption(dhcp.OptDomainServer, []byte(s.DNS))
		response.AddOption(dhcp.OptDomainName, []byte(s.DomainName))

		if pxeRequest {
			response.SetBootFile(pxe.BootFile)
			response.AddOption(dhcp.OptRootPath, []byte(pxe.RootPath))
		}

		//update lease information
		lease.Stat = ALLOCATED
		lease.Expiry = time.Now().Add(s.LeaseDuration)

		return response

	case dhcp.DECLINE:
		if serverId, ok := options[dhcp.OptDHCPServerId]; ok && !net.IP(serverId).Equal(s.Ip) {
			return nil
		}
		if lease == nil {
			return nil
		}
		//set the IP address to NOTAVAILABLE
		lease.Stat = NOTAVAILABLE
		lease.Mac = net.HardwareAddr{}
		newLease := pool.GetAvailLease()
		newLease.Stat = RESERVED
		newLease.Mac = mac
		//Update database
		//db.UpdateBindingTable(mac,newLease.Ip)
		return nil

	case dhcp.RELEASE:
		if serverId, ok := options[dhcp.OptDHCPServerId]; ok && !net.IP(serverId).Equal(s.Ip) {
			return nil
		}
		if lease == nil {
			return nil
		}
		//set the IP address to RESERVED
		lease.Stat = RESERVED
		return nil
	}

	return nil
}
