package pxedhcp

import (
	"net"
	"strconv"
	"clowder/pxedhcp"
)



func DHCPResponder(p pxedhcp.Packet, s *Server) pxedhcp.Packet {

	options := p.ParseOptions()
	if val, ok := options[pxedhcp.OptDHCPMsgType]; !ok || len(val)!=1 {
		return nil
	}
	msgType := val[0]
	if msgType < pxedhcp.DISCOVER || msgType > pxedhcp.INFORM {
		return nil
	}

	s.pool.Refresh()

	switch msgType {

	case pxedhcp.DISCOVER:

		if offerIP:=s.pool.GetIPFromMac(p.GetHardwareAddr()); offerIP==nil{
			if offerIP = s.pool.GetAvailIP(); offerIP ==nil {
				return nil
			}
		}

		response := pxedhcp.NewReplyPacket(p)
		//set packet header
		response.SetYIAddr(offerIP)
		response.SetServerName(s.ServerName)
//		response.SetBootFile(s.BootFile)
		//set packet options
		//MUST
		response.AddOPtion(pxedhcp.OptDHCPMsgType,pxedhcp.OFFER)	//Message Type
		response.AddOPtion(pxedhcp.OptAddressTime,s.leaseDuration)	//Lease time
		response.AddOption(pxedhcp.OptDHCPServerId,s.ip)		//Server identifier
		//MAY
		response.AddOPtion(pxedhcp.OptSubnetMask,s.mask)
		response.AddOption(pxedhcp.OptRouter,s.Router)
		response.AddOption(pxedhcp.OptDomainServer,s.DNS)
		response.AddOption(pxedhcp.OptDomainName,s.DomainName)

		if p.IsPxeRequest() {
			uuid:=options[pxedhcp.OptUUIDGUID]
			if pxeRecord:=s.PxeTable.GetRecord(uuid);pxeRecord != nil{
				response.SetBootFile(pxeRecord.BootFile)
				response.AddOption(pxedhcp.OptRootPath,pxeRecord.RootPath)
			} else {
				//TODO: WARNING : new machine
			}
		}

		//update leases
		if s.pool.GetIPStat(offerIP)==AVAILABLE {
			s.pool.SetIPStat(offerIP,RESERVED)
			s.pool.SetExpTime(offerIP, time.Now().Add(time.Minute*5))
		}
		return response

	case REQUEST:

		if serverId, ok:= options[OptDHCPServerId]; ok &&  !net.IP(erverId).Equal(s.ip){
			return nil
		}

		if offerIP:=s.pool.GetIPFromMac(p.GetHardwareAddr()); offerIP==nil{
			//TODO WARN no record of this client
			return nil

		}
		if requestIP := net.IP(options[OptAddressRequest]); requestIP==nil {
			requestIP= p.GetCIAddr()
		}

		response := pxedhcp.NewReplyPacket(p)

		if requestIP.Equal(net.IPv4Zero) || !requestIP.Equal(offerIP) { 
			response.AddOPtion(pxedhcp.OptDHCPMsgType,pxedhcp.NAK)
			return response
		}
		//set packet header
		response.SetYIAddr(IPOffer)
		response.SetServerName(s.ServerName)
//		response.SetBootFile(s.BootFile)
		//set packet options
		//MUST
		response.AddOPtion(pxedhcp.OptDHCPMsgType,pxedhcp.ACK)		//Message Type
		response.AddOPtion(pxedhcp.OptAddressTime,s.leaseDuration)	//Lease time
		response.AddOption(pxedhcp.OptDHCPServerId,s.ip)		//Server identifier
		//MAY
		response.AddOPtion(pxedhcp.OptSubnetMask,s.mask)
		response.AddOption(pxedhcp.OptRouter,s.Router)
		response.AddOption(pxedhcp.OptDomainServer,s.DNS)
		response.AddOption(pxedhcp.OptDomainName,s.DomainName)

		if p.IsPxeRequest() {
			uuid:=options[pxedhcp.OptUUIDGUID]
			if pxeRecord:=s.PxeTable.GetRecord(uuid);pxeRecord != nil{
				response.SetBootFile(pxeRecord.BootFile)
				response.AddOption(pxedhcp.OptRootPath,pxeRecord.RootPath)
			} else {
				//TODO: WARNING : new machine
			}
		}

		return response


	case RELEASE, DECLINE:
		//TODO : update leases
		return nil
}
