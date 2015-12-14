package server

import (
	"net"
	"log"
	"fmt"
	"strconv"
	"time"
	"clowder/pxedhcp"
//	"clowder/db"
	"clowder/dbase"
)
type Hardware struct {
	RequestTime	time.Time
	Uuid		string
}

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
	MachineLeases	dbase.Leases
	DeviceLeases	dbase.Leases
	Pxe		dbase.PxeTable
	NewHardware	dbase.Hardwares
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
func NewServer(ip, mask net.IP, port int, duration time.Duration, hostname string, dns, router net.IP, domainName string) *Server {
	s := new(Server)
	s.Ip = ip
	s.Mask = mask
	s.TcpPort = port
	s.LeaseDuration = duration
	s.ServerName = hostname
	if dns==nil { dns=ip }
	s.DNS = dns
	if router==nil {router=ip}
	s.Router = router
	s.DomainName=domainName
	s.NewHardware=make(dbase.Hardwares)

	s.TablesAccess = make(chan bool,1)
	s.TablesAccess<-true
	s.LogAccess = make(chan bool,1)
	s.LogAccess <-true
	s.DHCPOn = make(chan bool,1)
	s.DHCPOn <-false
	s.Logger=nil
	return s
}

//StartTCPServer run a TCP server
func (s *Server) StartTCPServer() error {
	service := ":"+strconv.Itoa(s.TcpPort)
	tcpAddr, err:= net.ResolveTCPAddr("tcp4",service)
	if err !=nil { return err  }
	listener, err:= net.ListenTCP("tcp4",tcpAddr)
	if err !=nil { return err }
	s.TcpConn = make([]*net.TCPConn,0,10)
	s.TcpQuit = make(chan bool)

	s.WriteLog("INFO\tClowder is running on port " + strconv.Itoa(s.TcpPort))


	for {
		conn, err := listener.AcceptTCP()
		if err !=nil {

			select {
			case <-s.TcpQuit:	//error happened because the listener is closed
				for i:= range s.TcpConn {
					if s.TcpConn[i]!=nil { s.TcpConn[i].Close() }
				}
				return nil
			default:		//other errors
			}
			continue
		}

		s.TcpConn=append(s.TcpConn,conn)
		//handle connection
		go func(conn *net.TCPConn){
			defer conn.Close()
			var buf [512]byte
			addr:=conn.RemoteAddr().String()
			s.WriteLog("INFO\tNew connection from "+conn.RemoteAddr().String())
			for {
				n, err:= conn.Read(buf[0:])
				if err!=nil {
					s.WriteLog("ERROR\t"+err.Error())
					return
				}
				if n<=2 { continue }
				cmd:=string(buf[:n])
				if cmd[n-2:]==string([]byte{13,10}){ cmd=cmd[:n-2] }
				s.WriteLog("INFO\tGet command "+string(cmd)+" from "+addr )
				switch cmd {
					case "DHCPON":
						go s.StartDHCPServer()
						conn.Write([]byte("DONE"))
						time.Sleep(time.Second*5)

					case "DHCPOFF":
						s.StopDHCPServer()
						conn.Write([]byte("DONE"))

					case "LEASES":
						msg:=s.ExportLeaseTable()
						fmt.Println(msg)
						conn.Write([]byte(msg))
					case "STOPCLOWDER":
						conn.Write([]byte("CLOWDER closing..."))
						on:=<-s.DHCPOn
						s.DHCPOn<-on
						if on { s.StopDHCPServer() }
						close(s.TcpQuit)
						listener.Close()
					case "NEWHARDWARE":
						msg:=s.NewHardware.String()
						fmt.Println(msg)
						conn.Write([]byte(msg))
					case "STATUS":
						msg:=s.GetStatus()
						fmt.Println(msg)
						conn.Write([]byte(msg))
					case "CLOSECONN":
						conn.Write([]byte("DONE"))
						return
					default:
						conn.Write([]byte("INVALID COMMAND.\nUSE: DHCPON, DHCPOFF, LEASES, NEWHARDWARE, STATUS, CLOSECONN, STOPCLOWDER"))
						continue
				}
			}
		}(conn)

	}
	return nil
}

func (s *Server) StartDHCPServer() {
	//check DHCP server status
	on:=<-s.DHCPOn
	if on {
		s.DHCPOn<-true
		s.WriteLog("ERROR\tTried to start a DHCP server which is running")
		return
	}

	udpAddr,_ := net.ResolveUDPAddr("udp4",":5001")
	conn, err := net.ListenUDP("udp4",udpAddr)
	if err != nil {
		s.DHCPOn<-false
		s.WriteLog("ERROR\t"+err.Error())
		return
	}

	s.UdpConn = conn
	s.DHCPOn<-true

	defer func(){
		<-s.DHCPOn
		s.DHCPOn<-false
	}()

	buffer := make([]byte, 1500)

	s.WriteLog("INFO\tDHCP server is enabled")

	for {
		n, err := s.UdpConn.Read(buffer)
		//check error, s.DHCPOn==false means the listener is closed by user
		if err != nil {
			on=<-s.DHCPOn
			s.DHCPOn<-on
			if on {
				s.WriteLog("ERROR\t"+err.Error())
			}
			return

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
			s.WriteLog("ERROR\t"+err.Error())
			return
		}
	}
}

func (s *Server) StopDHCPServer() {
	on:=<-s.DHCPOn
	if on {
		s.UdpConn.Close()
		s.WriteLog("INFO\tDHCP server is disabled")
	} else {
		s.WriteLog("ERROR\tTried to stop a DHCP server without being started")
	}
	s.DHCPOn<-false
}

func (s *Server) WriteLog(message string) {
	<-s.LogAccess
	s.Logger.Println(message)
	fmt.Println(message)
	s.LogAccess<-true
}

func (s *Server) ExportLeaseTable() string {
	//result:=""
	//if str:=s.MachineLeases.String();str!="" {result+=str+"\n"}
	//result+=s.DeviceLeases.String()
	//return result
	return s.MachineLeases.String()+"\n"+s.DeviceLeases.String()
}

func (s *Server) GetStatus() string {
	on:=<-s.DHCPOn
	<-s.TablesAccess
	msg:="Clowder server:\n\tIP Address: "+s.Ip.String()+"\n\tSubnet mask: "+s.Mask.String()+"\n\tPort: "+ strconv.Itoa(s.TcpPort)
	msg+="\nDHCP server is "
	if on {
		msg+="active."
	} else {
		msg+="inactive."
	}
	msg+="\nCurrent leases:\n"+s.MachineLeases.String()+"\n"+s.DeviceLeases.String()
	msg+="PXE Information:\n"+s.Pxe.String()
	s.DHCPOn<-on
	s.TablesAccess<-true
	return msg
}


func SendCommand(addr,cmd string) (string, error) {
	var (
		tcpAddr	*net.TCPAddr
		conn	*net.TCPConn
		err	error
		buf	[2048]byte
		n	int
	)
	if tcpAddr, err=net.ResolveTCPAddr("tcp4",addr); err!=nil { return "", err }
	if conn, err = net.DialTCP("tcp4",nil,tcpAddr); err!=nil { return "", err }
	if _, err = conn.Write([]byte(cmd)); err!=nil { return "", err }
	if n,err = conn.Read(buf[0:]); err!=nil { return "", err }
	if cmd!="CLOSECONN" {
		if _, err = conn.Write([]byte("CLOSECONN")); err!=nil {
			return "", err
		}
	}
	return string(buf[:n]),nil
}

