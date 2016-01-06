package cmd

import (
	"github.com/spf13/cobra"
)

var tcpPort int

var RootCmd = &cobra.Command{Use: "clowder"}

func init() {
	RootCmd.PersistentFlags().IntVarP(&tcpPort, "port", "p", 5000, "TCP port to connect")

}
