package server

import (
	"net"
//	"strconv"
	"clowder/pxedhcp"
	"time"
	"encoding/binary"
)



func (s *Server) DHCPResponder(p pxedhcp.Packet) pxedhcp.Packet {

	//Get lock
	<-s.TablesAccess

	defer func() {
		s.TablesAccess<-true
	}()
	options := p.ParseOptions()

	//Check DHCP Type
	val, ok := options[pxedhcp.OptDHCPMsgType]
	if !ok || len(val)!=1 {
		return nil
	}
	msgType := val[0]
	if msgType < pxedhcp.DISCOVER || msgType > pxedhcp.INFORM {
		return nil
	}
	//Is a PXE DHCP Packet
	pxeRequest := options[pxedhcp.OptClassId]!=nil &&  options[pxedhcp.OptClientSystemArchitecture]!=nil && options[pxedhcp.OptClientNetworkDeviceInterface]!=nil && options[pxedhcp.OptUUIDGUID]!=nil
	var pool Leases
	var uuid []byte
	if pxeRequest {
		uuid=options[pxedhcp.OptUUIDGUID]
		pool=s.MachineLeases
	} else {
		pool=s.DeviceLeases
	}

	pool.Refresh()
	duration := make([]byte,4)
	binary.BigEndian.PutUint32(duration,uint32(s.LeaseDuration/time.Second))


	switch msgType {

	case pxedhcp.DISCOVER:
		lease:=pool.GetLeaseFromMac(p.GetHardwareAddr())
		if  lease==nil{
			if lease = pool.GetAvailLease(); lease ==nil {
				return nil
			}
		}

		response := pxedhcp.NewReplyPacket(p)
		//set packet header
		response.SetYIAddr(lease.ip)
		response.SetServerName(s.ServerName)
//		response.SetBootFile(s.BootFile)
		//set packet options
		//MUST
		response.AddOption(pxedhcp.OptDHCPMsgType,[]byte{pxedhcp.OFFER})	//Message Type
		response.AddOption(pxedhcp.OptAddressTime,duration)	//Lease time
		response.AddOption(pxedhcp.OptDHCPServerId,s.Ip)		//Server identifier
		//MAY
		response.AddOption(pxedhcp.OptSubnetMask,s.Mask)
		response.AddOption(pxedhcp.OptRouter,s.Router)
		response.AddOption(pxedhcp.OptDomainServer,s.DNS)
		response.AddOption(pxedhcp.OptDomainName,[]byte(s.DomainName))

		if pxeRequest {
		//i	uuid:=options[pxedhcp.OptUUIDGUID]
			if pxeRecord:=s.Pxe.GetRecord(uuid);pxeRecord != nil{
				response.SetBootFile(pxeRecord.BootFile)
				response.AddOption(pxedhcp.OptRootPath,[]byte(pxeRecord.RootPath))
			} else {
				//TODO: WARNING : new machine
			}
		}

		//update leases

		if lease.stat==AVAILABLE {
			lease.stat=RESERVED
			lease.expiry=time.Now().Add(time.Minute*5)
		}
		return response

	case pxedhcp.REQUEST:

		if serverId, ok:= options[pxedhcp.OptDHCPServerId]; ok &&  !net.IP(serverId).Equal(s.Ip){
			return nil
		}
		lease:=pool.GetLeaseFromMac(p.GetHardwareAddr())
		if lease==nil{
			//TODO WARN no record of this client
			return nil

		}
		requestIP := net.IP(options[pxedhcp.OptAddressRequest])
		if requestIP==nil {
			requestIP= p.GetCIAddr()
		}

		response := pxedhcp.NewReplyPacket(p)

		if requestIP.Equal(net.IPv4zero) || !requestIP.Equal(lease.ip) {
			response.AddOption(pxedhcp.OptDHCPMsgType,[]byte{pxedhcp.NAK})
			return response
		}
		//set packet header
		response.SetYIAddr(lease.ip)
		response.SetServerName(s.ServerName)
//		response.SetBootFile(s.BootFile)
		//set packet options
		//MUST
		response.AddOption(pxedhcp.OptDHCPMsgType,[]byte{pxedhcp.ACK})		//Message Type
		response.AddOption(pxedhcp.OptAddressTime,duration)	//Lease time
		response.AddOption(pxedhcp.OptDHCPServerId,s.Ip)		//Server identifier
		//MAY
		response.AddOption(pxedhcp.OptSubnetMask,s.Mask)
		response.AddOption(pxedhcp.OptRouter,s.Router)
		response.AddOption(pxedhcp.OptDomainServer,s.DNS)
		response.AddOption(pxedhcp.OptDomainName,[]byte(s.DomainName))

		if pxeRequest {
			//uuid:=options[pxedhcp.OptUUIDGUID]
			if pxeRecord:=s.Pxe.GetRecord(uuid);pxeRecord != nil{
				response.SetBootFile(pxeRecord.BootFile)
				response.AddOption(pxedhcp.OptRootPath,[]byte(pxeRecord.RootPath))
			} else {
				//TODO: WARNING : new machine
			}
		}

		return response


	case pxedhcp.RELEASE, pxedhcp.DECLINE:
		//TODO : update leases
		return nil
	}

	return nil
}
