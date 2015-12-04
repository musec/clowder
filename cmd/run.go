package cmd

import (
        "github.com/spf13/cobra"
	"clowder/server"
        "fmt"
	"os"
	"github.com/spf13/viper"
	"net"
	"log"
)

var	tcpPort	int

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Clowder server management service.",
	Long:  `Start Clowder server management service.
	`,
	Run: runRun,
}

func runRun(cmd *cobra.Command, args []string) {
	//Read config file
	viper.SetConfigType("toml")
	viper.SetConfigName("clowder_config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if tcpPort<1024 || tcpPort > 65535 {
		panic(fmt.Errorf("Cannot use port %d. TCP port must be a registered port.", tcpPort))
	}

	//Create server
	fmt.Println("Starting Clowder...")
	serverIP := net.ParseIP(viper.GetString("server.ip"))
	serverMask :=  net.ParseIP(viper.GetString("server.subnetmask"))
	s := server.NewServer(serverIP,serverMask,tcpPort)

	//Create log file
	logFile:="clowder.log"
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to open log file", logFile, ":", err)
		file=os.Stdout
	}
	s.Logger=log.New(file,"",log.Ldate|log.Ltime)

	//Setup machine IP pool
	machineIP:=net.ParseIP(viper.GetString("machines.ipstart"))
	machineRange:=viper.GetInt("machines.iprange")
	s.MachineLeases = server.NewLeases(machineIP,machineRange)

	//Setup device IP pool
	deviceIP:=net.ParseIP(viper.GetString("devices.ipstart"))
	deviceRange:=viper.GetInt("devices.iprange")
	s.DeviceLeases = server.NewLeases(deviceIP,deviceRange)


	if err := s.StartTCPServer(); err!=nil {
		panic(err)
	}

}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVarP(&tcpPort, "port", "n", 5000, "TCP port to connect")
}
