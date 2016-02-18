package cmd

import (
	"github.com/musec/clowder/server"
	"github.com/spf13/cobra"
)

var dhcpCmd = &cobra.Command{Use: "dhcp"}

func init() {
	RootCmd.AddCommand(dhcpCmd)

	dhcpCmd.AddCommand(dhcpOnCmd)
	dhcpCmd.AddCommand(dhcpOffCmd)
	dhcpCmd.AddCommand(statCmd)
}

var dhcpOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Enable DHCP service",
	Long: `Enable DHCP service.
	`,
	Run: dhcpOnRun,
}

var dhcpOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Disable DHCP service",
	Long: `Disable DHCP service.
	`,
	Run: dhcpOffRun,
}

var statCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Clowder's status",
	Long: `Show Clowder's status".
	Show current leases, new machines and new devices.`,
	Run: statRun,
}

func dhcpOnRun(cmd *cobra.Command, args []string) {
	c := server.ConnectOrDie(config)

	if msg, err := c.SendCommand("DHCPON"); err == nil {
		c.Log(msg)
	} else {
		c.FatalError(err)
	}

}

func dhcpOffRun(cmd *cobra.Command, args []string) {
	c := server.ConnectOrDie(config)

	if msg, err := c.SendCommand("DHCPOFF"); err == nil {
		c.Log(msg)
	} else {
		c.FatalError(err)
	}

}

func statRun(cmd *cobra.Command, args []string) {
	c := server.ConnectOrDie(config)

	if msg, err := c.SendCommand("STATUS"); err == nil {
		c.Log(msg)
	} else {
		c.FatalError(err)
	}

}
