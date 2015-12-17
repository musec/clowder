package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
	"github.com/musec/clowder/server"
	"strconv"
)
var tcpAddr string

var dhcpCmd = &cobra.Command{Use:"dhcp"}

func init() {
	RootCmd.AddCommand(dhcpCmd)
	dhcpCmd.PersistentFlags().StringVarP(&tcpAddr, "addr", "a", "localhost", "IP Address of Clowder server")
	dhcpCmd.AddCommand(dhcpOnCmd)
	dhcpCmd.AddCommand(dhcpOffCmd)
	dhcpCmd.AddCommand(statCmd)

}

var dhcpOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Enable DHCP service",
	Long:  `Enable DHCP service.
	`,
	Run: dhcpOnRun,
}

var dhcpOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Disable DHCP service",
	Long:  `Disable DHCP service.
	`,
	Run: dhcpOffRun,
}


var statCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Clowder's status",
	Long:  `Show Clowder's status".
	Show current leases, new machines and new devices.`,
	Run: statRun,
}

func dhcpOnRun(cmd *cobra.Command, args []string) {
	addr:=tcpAddr+":"+strconv.Itoa(tcpPort)
	fmt.Println("Connected to ",addr)

	if msg,err:=server.SendCommand(addr,"DHCPON"); err==nil {
		fmt.Println(msg)
	} else {
		fmt.Println(err.Error())
	}

}

func dhcpOffRun(cmd *cobra.Command, args []string) {
	addr:=tcpAddr+":"+strconv.Itoa(tcpPort)
	if msg,err:=server.SendCommand(addr,"DHCPOFF"); err==nil {
		fmt.Println(msg)
	} else {
		fmt.Println(err.Error())
	}

}

func statRun(cmd *cobra.Command, args []string) {
	addr:=tcpAddr+":"+strconv.Itoa(tcpPort)
	if msg,err:=server.SendCommand(addr,"STATUS"); err==nil {
		fmt.Println(msg)
	} else {
		fmt.Println(err.Error())
	}

}


