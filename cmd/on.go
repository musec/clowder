package commands

import (
        "github.com/spf13/cobra"
        "fmt"
)

var dhcpOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Enable DHCP service",
	Long:  `Start Clowder server management service.
	`,
	Run: dhcpOnRun,
}

func dhcpOnRun(cmd *cobra.Command, args []string) {
	fmt.Println("DHCP on")
}

func init() {
	dhcpCmd.AddCommand(dhcpOnCmd)
}
