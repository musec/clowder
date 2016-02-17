package server

import (
	"fmt"
	"github.com/spf13/viper"
	"net"
	"os"
)

//
// A connection to a server (wraps an internal TCP connection).
//
type Connection struct {
	connection net.Conn
}

// Connect to a server named in a Viper configuration.
func Connect(config *viper.Viper) (*Connection, error) {
	host := config.GetString("server.host")
	port := config.GetString("server.controlPort")

	server := host + ":" + port

	connection, err := net.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	return &Connection{connection}, nil
}

// Connect to a server or terminate the application.
func ConnectOrDie(config *viper.Viper) *Connection {
	c, err := Connect(config)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return c
}

func (c Connection) SendCommand(cmd string) (string, error) {
	var buffer [2048]byte

	if _, err := c.connection.Write([]byte(cmd)); err != nil {
		return "", err
	}

	n, err := c.connection.Read(buffer[0:])
	if err != nil {
		return "", err
	}

	if cmd != "CLOSECONN" {
		if _, err := c.connection.Write([]byte("CLOSECONN")); err != nil {
			return "", err
		}
	}
	return string(buffer[:n]), nil
}
