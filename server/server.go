package server

import (
	"net"
	"log"
	"fmt"
	"strconv"
	"time"
	"clowder/pxedhcp"
//	"clowder/db"

)

type Server struct {
	//Server information
	Ip		net.IP
	Mask		net.IP
	TcpPort		int
	LeaseDuration	time.Duration
	ServerName	string
	DNS		net.IP
	Router		net.IP
	DomainName	string

	//Leases management
	MachineLeases	Leases
	DeviceLeases	Leases
	Pxe		PxeTable
	TablesAccess	chan bool

	//connections
	TcpConn		[]*net.TCPConn
	UdpConn		*net.UDPConn
	TcpQuit		chan bool
	DHCPOn		chan bool

	//Logging
	Logger		*log.Logger
	LogAccess	chan bool
}

//NewServer 
func NewServer(ip_, mask_ net.IP, port int) *Server {
	s := new(Server)
	s.Ip = ip_
	s.Mask = mask_
	s.TcpPort = port
	s.TablesAccess = make(chan bool,1)
	s.TablesAccess<-true
	s.LogAccess = make(chan bool,1)
	s.LogAccess <-true
	s.DHCPOn = make(chan bool,1)
	s.DHCPOn <-false
	return s
}

//StartTCPServer run a TCP server
func (s *Server) StartTCPServer() error {
	service := ":"+strconv.Itoa(s.TcpPort)
	tcpAddr, err:= net.ResolveTCPAddr("tcp4",service)
	if err !=nil {
		return err
	}
	listener, err:= net.ListenTCP("tcp4",tcpAddr)
	if err !=nil {
		return err
	}
	s.TcpConn = make([]*net.TCPConn,0,10)
	s.TcpQuit = make(chan bool)

	for {
		conn, err := listener.AcceptTCP()
		if err !=nil {

			select {
			case <-s.TcpQuit:
				fmt.Println("Clowder closed.")
				return nil
			default:
			}
			continue
		}
		s.TcpConn=append(s.TcpConn,conn)
		//handleClient(conn)
		go func(conn *net.TCPConn){
			defer conn.Close()
			var buf [64]byte
			for {
				n, err:= conn.Read(buf[0:])
				if err!=nil {
					return
				}
				if n<=2 {
					continue
				}
				msg:=string(buf[:n-2])
				fmt.Printf( "Get message: %s \n", msg)
				switch msg {
					case "DHCPON":
						s.StartDHCPServer()
					case "DHCPOFF":
						s.StopDHCPServer()
					case "LEASES":
						fmt.Println([]byte(s.ExportLeaseTable()))
						//conn.Write([]byte(s.ExportLeaseTable()))
					case "PXE":
						fmt.Println([]byte(s.Pxe.Export()))
						//conn.Write([]byte(s.Pxe.Export()))
					case "QUIT":
						s.StopDHCPServer()
						close(s.TcpQuit)
						listener.Close()
					case "NEWMACHINE":
					case "exit":
						return
					default:
						continue
				}
			}
		}(conn)

	}
	return nil
}

func (s *Server) StartDHCPServer() error {
	var err error

	udpAddr, err := net.ResolveUDPAddr("udp4",":5001")
	if err!=nil{
		return err
	}

	conn, err := net.ListenUDP("udp4",udpAddr)
	if err != nil {
		return err
	}

	s.UdpConn = conn

	defer s.UdpConn.Close()

	buffer := make([]byte, 1500)

	log.Println("Trace: DHCP Server Listening.")

	for {

		n, err := s.UdpConn.Read(buffer)
		if err != nil {
			return err
		}
		if n < 240 {
			continue
		}

		requestPacket := pxedhcp.Packet(buffer[:n])

		responsePacket := s.DHCPResponder(requestPacket)

		if responsePacket == nil {
			continue
		}
		if _, err := conn.WriteTo(responsePacket, &net.UDPAddr{IP: net.IPv4bcast, Port: 68}); err != nil {
			return err
		}
	}
}

func (s *Server) StopDHCPServer() {
	on:=<-s.DHCPOn
	if on {
		s.UdpConn.Close()
	}
	s.DHCPOn<-false
}

func (s *Server) WriteLog(logType, message string) {
	<-s.LogAccess
	s.Logger.Println(logType,": ", message)
	s.LogAccess<-true
}

func (s *Server) ExportLeaseTable() string {
	return s.MachineLeases.Export()+"\n"+s.DeviceLeases.Export()
}
/*
func handleClient(conn *net.TCPConn){
	defer conn.Close()

	var buf [64]byte
	n, err:= conn.Read(buf[0:])
	if err!=nil {
		return
	}

	fmt.Printf( "Get message: %s \n", buf[0:n])
	return
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	}

}*/
