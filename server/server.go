package server

import (
	"net"
	"log"
	"fmt"
	"strconv"
	"time"
	"clowder/pxedhcp"
	"clowder/db"

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
	MachinePool	*IPPool
	DevicePool	*IPPool
	Pxe		*PxeTable
	TableAccess	chan bool

	//connections
	TcpConn		[]*net.TCPConn
	UdpConn		*net.UDPConn
	TcpQuit		chan bool
	UDPServerAccess	chan bool

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
	s.TableAccess = make(chan bool,1)
	s.TableAccess<-true
	s.LogAccess = make(chan bool,1)
	s.LogAccess <-true
	s.UDPServerAccess = make(chan bool,1)
	s.UDPServerAccess <-false
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
		s.TcpConn.append(conn)
		//handleClient(conn)
		go func(c *net.TCPConn){
			defer c.Close()
			var buf [64]byte
			for {
				n, err:= s.tcpConn.Read(buf[0:])
				if err!=nil || n<=2 {
					return
				}
				msg:=string(buf[:n-2])
				fmt.Printf( "Get message: %s \n", msg)
				switch msg {
					case "DHCPON":
					case "DHCPOFF":
					case "LEASES":
					case "QUIT":
						close(s.
						close(s.tcpQuit)
						listener.Close()
					case "PXE":
					case "NEWMACHINE":
					case "exit":
				}
			}
		}(s.tcpConn)

	}
	return nil
}

func (s *Server) StartDHCPServer() error {
	var err error

	s.udpConn, err := net.ListenUDP("udp4",":67")
	if err != nil {
		return err
	}

	defer s.udpConn.Close()

	buffer := make([]byte, 1500)

	log.Println("Trace: DHCP Server Listening.")

	for {
		select {
		case <-s.udpQuit:
			fmt.Println("DHCP server off closed.")
			return nil
		default:
		}

		n, err := s.udpConn.Read(buffer)
		if err != nil {
			return err
		}
		if n < 240 {
			continue
		}

		requestPacket := Packet(buffer[:n])

		responsePacket := pxedhcp.DHCPResponder(requestPacket, s)

		if responePacket == nil {
			continue
		}
		if _, err := conn.WriteTo(responsePacket, &net.UDPAddr{IP: net.IPv4bcast, Port: 68}); err != nil {
			return err
		}
	}
}


func (s *Server) WriteLog(logType, message : string) {
	<-s.LogAccess
	s.Logger.Println(logType,": ", message)
	s.LogAccess<-true
}

func (s *Server) ExportLeaseTable() string {
	result:=""
	for l:=range s.MachinePool.Leases{
		
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
