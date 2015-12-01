package commands

import (
	"github.com/spf13/cobra"
)

var dhcpCmd = &cobra.Command{Use:"dhcp"}

func init() {
	RootCmd.AddCommand(dhcpCmd)
}
