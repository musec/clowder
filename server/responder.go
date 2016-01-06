package server

import (
	"encoding/binary"
	"github.com/musec/clowder/dbase"
	"github.com/musec/clowder/pxedhcp"
	"net"
	"time"
)

func (s *Server) DHCPResponder(p pxedhcp.Packet) pxedhcp.Packet {

	//Get lock
	<-s.TablesAccess
	defer func() { s.TablesAccess <- true }()
	options := p.ParseOptions()

	//Get DHCP Message Type
	val, ok := options[pxedhcp.OptDHCPMsgType]
	if !ok || len(val) != 1 {
		return nil
	}
	msgType := val[0]
	if msgType < pxedhcp.DISCOVER || msgType > pxedhcp.INFORM {
		return nil
	}

	//Is a PXE DHCP Packet
	pxeRequest := options[pxedhcp.OptClassId] != nil && options[pxedhcp.OptClientSystemArchitecture] != nil && options[pxedhcp.OptClientNetworkDeviceInterface] != nil && options[pxedhcp.OptUUIDGUID] != nil
	var pool dbase.Leases
	var uuid dbase.UUID
	var pxe *dbase.PxeRecord
	if pxeRequest {
		if len(options[pxedhcp.OptUUIDGUID]) == 17 {
			uuid = options[pxedhcp.OptUUIDGUID][1:]
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
	mac := p.GetHardwareAddr()
	lease := pool.GetLeaseFromMac(mac)

	switch msgType {
	case pxedhcp.DISCOVER:
		s.WriteLog("INFO\tGet DISCOVER message from " + mac.String())
		if lease == nil { //no record of this MAC address
			s.NewHardware[mac.String()] = uuid
			return nil
		}

		response := pxedhcp.NewReplyPacket(p)
		//set packet header
		response.SetYIAddr(lease.Ip)
		response.SetServerName(s.ServerName)
		//set packet options
		//MUST
		response.AddOption(pxedhcp.OptDHCPMsgType, []byte{pxedhcp.OFFER}) //Message Type
		response.AddOption(pxedhcp.OptAddressTime, duration)              //Lease time
		response.AddOption(pxedhcp.OptDHCPServerId, []byte(s.Ip))         //Server identifier
		//MAY
		response.AddOption(pxedhcp.OptSubnetMask, []byte(s.Mask))
		response.AddOption(pxedhcp.OptRouter, []byte(s.Router))
		response.AddOption(pxedhcp.OptDomainServer, []byte(s.DNS))
		response.AddOption(pxedhcp.OptDomainName, []byte(s.DomainName))

		if pxeRequest {
			if pxe != nil {
				response.SetBootFile(pxe.BootFile)
				response.AddOption(pxedhcp.OptRootPath, []byte(pxe.RootPath))
			} else {
				s.WriteLog("WARN Receive PXE DHCP DISCOVER packet from a known MAC address: " + mac.String() + ". But there isn't PXE information to response.")
				return nil
			}
		}

		//update leases. UPDATE: dont have to update since DHCP server only response to packet from binded MAC address
		//if lease.stat==AVAILABLE {
		//	lease.stat=RESERVED
		//	lease.expiry=time.Now().Add(time.Minute*5)
		//}

		return response

	case pxedhcp.REQUEST:
		s.WriteLog("INFO\tGet REQUEST message from " + mac.String())
		//Is the packet for this server
		if serverId, ok := options[pxedhcp.OptDHCPServerId]; ok && !net.IP(serverId).Equal(s.Ip) {
			return nil
		}

		if lease == nil { //no record of this MAC address
			//s.NewMachines=append(s.NewMachines,Machines{mac,time.Now(),pxeRequest})
			s.WriteLog("WARN Receive a DHCP REQUEST packet from an unknown MAC address: " + mac.String() + ".")
			return nil
		}

		requestIP := net.IP(options[pxedhcp.OptAddressRequest])
		if requestIP == nil { //the request packet is for extending its lease
			requestIP = p.GetCIAddr()
		}

		response := pxedhcp.NewReplyPacket(p)

		//If the request IP is different the offer IP
		if requestIP.Equal(net.IPv4zero) || !requestIP.Equal(lease.Ip) {
			s.WriteLog("Request IP " + requestIP.String() + " is different to offer IP " + lease.Ip.String())
			response.AddOption(pxedhcp.OptDHCPMsgType, []byte{pxedhcp.NAK})
			return response
		}

		if pxeRequest && pxe == nil {
			s.WriteLog("WARN Receive PXE DHCP REQUEST packet from a known MAC address: " + mac.String() + ". But there isn't PXE information to response.")
			response.AddOption(pxedhcp.OptDHCPMsgType, []byte{pxedhcp.NAK})
			return response
		}

		//set packet header
		response.SetYIAddr(lease.Ip)
		response.SetServerName(s.ServerName)
		//set packet options
		//MUST
		response.AddOption(pxedhcp.OptDHCPMsgType, []byte{pxedhcp.ACK}) //Message Type
		response.AddOption(pxedhcp.OptAddressTime, duration)            //Lease time
		response.AddOption(pxedhcp.OptDHCPServerId, []byte(s.Ip))       //Server identifier
		//MAY
		response.AddOption(pxedhcp.OptSubnetMask, []byte(s.Mask))
		response.AddOption(pxedhcp.OptRouter, []byte(s.Router))
		response.AddOption(pxedhcp.OptDomainServer, []byte(s.DNS))
		response.AddOption(pxedhcp.OptDomainName, []byte(s.DomainName))

		if pxeRequest {
			response.SetBootFile(pxe.BootFile)
			response.AddOption(pxedhcp.OptRootPath, []byte(pxe.RootPath))
		}

		//update lease information
		lease.Stat = dbase.ALLOCATED
		lease.Expiry = time.Now().Add(s.LeaseDuration)

		return response

	case pxedhcp.DECLINE:
		if serverId, ok := options[pxedhcp.OptDHCPServerId]; ok && !net.IP(serverId).Equal(s.Ip) {
			return nil
		}
		if lease == nil {
			return nil
		}
		//set the IP address to NOTAVAILABLE
		lease.Stat = dbase.NOTAVAILABLE
		lease.Mac = net.HardwareAddr{}
		newLease := pool.GetAvailLease()
		newLease.Stat = dbase.RESERVED
		newLease.Mac = mac
		//Update database
		//db.UpdateBindingTable(mac,newLease.Ip)
		return nil

	case pxedhcp.RELEASE:
		if serverId, ok := options[pxedhcp.OptDHCPServerId]; ok && !net.IP(serverId).Equal(s.Ip) {
			return nil
		}
		if lease == nil {
			return nil
		}
		//set the IP address to RESERVED
		lease.Stat = dbase.RESERVED
		return nil
	}

	return nil
}
