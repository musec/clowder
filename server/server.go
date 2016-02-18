package server

import (
	"database/sql"
	"github.com/musec/clowder/dbase"
	"github.com/musec/clowder/pxedhcp"
	"github.com/spf13/viper"
	"net"
	"os"
	"strconv"
	"time"
)

type Hardware struct {
	RequestTime time.Time
	Uuid        string
}

type Server struct {
	HasLogger

	//Server information
	Ip            net.IP
	Mask          net.IP
	controlPort   int
	LeaseDuration time.Duration
	ServerName    string
	DNS           net.IP
	Router        net.IP
	DomainName    string

	//Data management
	MachineLeases dbase.Leases
	DeviceLeases  dbase.Leases
	Pxe           dbase.PxeTable
	NewHardware   dbase.Hardwares
	DBase         *sql.DB
	TablesAccess  chan bool

	//connections
	tcpConn []*net.TCPConn
	udpConn *net.UDPConn
	tcpQuit chan bool
	dhcpOn  chan bool
}

// Create a new Clowder server using Viper-provided configuration values.
func New(config *viper.Viper) (*Server, error) {
	s := new(Server)
	var err error = nil

	ip := config.GetString("server.ip")

	config.SetDefault("server.dns", ip)
	config.SetDefault("server.router", ip)

	err = s.InitLog(config.GetString("server.log"))
	if err != nil {
		return nil, err
	}

	s.Ip = net.ParseIP(config.GetString("server.ip")).To4()
	s.Mask = net.ParseIP(config.GetString("server.subnetmask")).To4()
	s.controlPort = config.GetInt("server.controlPort")
	s.LeaseDuration = config.GetDuration("server.duration")

	s.ServerName, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	s.DNS = net.ParseIP(config.GetString("server.dns")).To4()
	s.Router = net.ParseIP(config.GetString("server.router")).To4()
	s.DomainName = config.GetString("server.domainname")

	s.NewHardware = make(dbase.Hardwares)
	s.Pxe = make(dbase.PxeTable, 0, 10)

	s.TablesAccess = make(chan bool, 1)
	s.TablesAccess <- true
	s.dhcpOn = make(chan bool, 1)
	s.dhcpOn <- false

	return s, err
}

//StartTCPServer run a TCP server
func (s *Server) StartTCPServer() error {
	service := ":" + strconv.Itoa(s.controlPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		return err
	}
	s.tcpConn = make([]*net.TCPConn, 0, 10)
	s.tcpQuit = make(chan bool)

	s.Log("Clowder is running on port " + strconv.Itoa(s.controlPort))

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {

			select {
			case <-s.tcpQuit: //error happened because the listener is closed
				for i := range s.tcpConn {
					if s.tcpConn[i] != nil {
						s.tcpConn[i].Close()
					}
				}
				return nil
			default: //other errors
			}
			continue
		}

		s.tcpConn = append(s.tcpConn, conn)
		//handle connection
		go func(conn *net.TCPConn) {
			defer conn.Close()
			var buf [512]byte
			addr := conn.RemoteAddr().String()
			s.Log("New connection from " + conn.RemoteAddr().String())
			for {
				n, err := conn.Read(buf[0:])
				if err != nil {
					s.Error("" + err.Error())
					return
				}
				if n <= 2 {
					continue
				}
				cmd := string(buf[:n])
				if cmd[n-2:] == string([]byte{13, 10}) {
					cmd = cmd[:n-2]
				}
				s.Log("Get command " + string(cmd) + " from " + addr)
				switch cmd {
				case "DHCPON":
					go s.StartDHCPServer()
					conn.Write([]byte("DONE"))
					time.Sleep(time.Second * 5)

				case "DHCPOFF":
					s.StopDHCPServer()
					conn.Write([]byte("DONE"))

				case "LEASES":
					msg := s.ExportLeaseTable()
					s.Log("Lease table:", msg)
					conn.Write([]byte(msg))
				case "STOPCLOWDER":
					conn.Write([]byte("CLOWDER closing..."))
					on := <-s.dhcpOn
					s.dhcpOn <- on
					if on {
						s.StopDHCPServer()
					}
					close(s.tcpQuit)
					listener.Close()
				case "NEWHARDWARE":
					msg := s.NewHardware.String()
					s.Log("New hardware:", msg)
					conn.Write([]byte(msg))
				case "STATUS":
					msg := s.GetStatus()
					s.Log("Status:", msg)
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
	on := <-s.dhcpOn
	if on {
		s.dhcpOn <- true
		s.Error("Tried to start a DHCP server which is running")
		return
	}

	udpAddr, _ := net.ResolveUDPAddr("udp4", ":67")
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		s.dhcpOn <- false
		s.Error("" + err.Error())
		return
	}

	s.udpConn = conn
	s.dhcpOn <- true

	defer func() {
		<-s.dhcpOn
		s.dhcpOn <- false
	}()

	buffer := make([]byte, 1500)

	s.Log("DHCP server is enabled")

	for {
		n, addr, err := s.udpConn.ReadFrom(buffer)
		//check error, s.dhcpOn==false means the listener is closed by user
		if err != nil {
			on = <-s.dhcpOn
			s.dhcpOn <- on
			if on {
				s.Error("" + err.Error())
			}
			return

		}
		if n < 240 {
			continue
		}
		ipStr, portStr, err := net.SplitHostPort(addr.String())
		if err != nil {
			s.Error("" + err.Error())
			return
		}

		requestPacket := pxedhcp.Packet(buffer[:n])
		responsePacket := s.DHCPResponder(requestPacket)
		if responsePacket == nil {
			continue
		}

		if net.ParseIP(ipStr).Equal(net.IPv4zero) || responsePacket.IsBroadcast() {
			port, _ := strconv.Atoi(portStr)
			addr = &net.UDPAddr{IP: net.IPv4bcast, Port: port}
		}
		if _, err := conn.WriteTo(responsePacket, addr); err != nil {
			s.Error("" + err.Error())
			return
		}
		//if _, err := conn.WriteTo(responsePacket, &net.UDPAddr{IP: net.IPv4bcast, Port: 68}); err != nil {
		//	s.Error(""+err.Error())
		//	return
		//}
	}
}

func (s *Server) StopDHCPServer() {
	on := <-s.dhcpOn
	if on {
		s.udpConn.Close()
		s.Log("DHCP server is disabled")
	} else {
		s.Error("Tried to stop a DHCP server without being started")
	}
	s.dhcpOn <- false
}

func (s *Server) ExportLeaseTable() string {
	//result:=""
	//if str:=s.MachineLeases.String();str!="" {result+=str+"\n"}
	//result+=s.DeviceLeases.String()
	//return result
	return s.MachineLeases.String() + "\n" + s.DeviceLeases.String()
}

func (s *Server) GetStatus() string {
	on := <-s.dhcpOn
	<-s.TablesAccess
	msg := "Clowder server:\n\tHostname: " + s.ServerName + "\n\tIP Address: " + s.Ip.String() + "\n\tSubnet mask: " + s.Mask.String() + "\n\tPort: " + strconv.Itoa(s.controlPort) + "\n\tDNS: " + s.DNS.String()
	msg += "\nDHCP server is "
	if on {
		msg += "active."
	} else {
		msg += "inactive."
	}
	msg += "\nCurrent leases(IP, Status, MAC, Expiry):\n" + s.MachineLeases.String() + "\n" + s.DeviceLeases.String()
	msg += "\nPXE Information(UUID, RootPath, BootFile):\n" + s.Pxe.String()
	s.dhcpOn <- on
	s.TablesAccess <- true
	return msg
}
