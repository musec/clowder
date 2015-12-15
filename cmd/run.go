package cmd

import (
        "github.com/spf13/cobra"
	"clowder/server"
	"clowder/dbase"
        "fmt"
	"os"
	"github.com/spf13/viper"
	"net"
	"log"
	"strconv"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Clowder server management service.",
	Long:  `Start Clowder server management service.
	`,
	Run: runRun,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Close Clowder server management service.",
	Long:  `Close Clowder server management service.
	`,
	Run: stopRun,
}


func runRun(cmd *cobra.Command, args []string) {
	//Read config file
	viper.SetConfigType("toml")
	viper.SetConfigName("clowder_config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		fmt.Println("Fatal error config file: ", err.Error())
	}
//	if tcpPort<1024 || tcpPort > 65535 {
//		panic(fmt.Errorf("Cannot use port %d. TCP port must be a registered port.", tcpPort))
//	}

	//Create server
	fmt.Println("Starting Clowder...")
	serverIP := net.ParseIP(viper.GetString("server.ip")).To4()
	serverMask :=  net.ParseIP(viper.GetString("server.subnetmask")).To4()
	duration := viper.GetDuration("server.duration")
	hostname,_ := os.Hostname();
	dns := net.ParseIP(viper.GetString("server.dns")).To4()
	router := net.ParseIP(viper.GetString("server.router")).To4()
	domainName := viper.GetString("server.domainname")
	dbFile := viper.GetString("database.filename")
	s := server.NewServer(serverIP, serverMask, tcpPort, duration, hostname, dns, router, domainName)

	//Create log file
	logFile:="clowder.log"
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to open log file", logFile, ":", err)
		file=os.Stdout
	}
	s.Logger=log.New(file,"",log.Ldate|log.Ltime)

	//Open databse
	if db,err:=sql.Open("sqlite3", dbFile); err==nil && db!=nil {
		s.DBase=db
	} else {
		s.WriteLog("ERROR\t"+err.Error())
		os.Exit(1)
	}

	//Setup machine IP pool
	machineIP:=net.ParseIP(viper.GetString("machines.ipstart"))
	machineRange:=viper.GetInt("machines.iprange")
	s.MachineLeases = dbase.NewLeases(machineIP,machineRange)
	if err:=s.MachineLeases.ReadBindingFromDB(s.DBase);err!=nil{
		s.WriteLog("ERROR\t"+err.Error())
		os.Exit(1)
	}

	//Setup device IP pool
	deviceIP:=net.ParseIP(viper.GetString("devices.ipstart"))
	deviceRange:=viper.GetInt("devices.iprange")
	s.DeviceLeases = dbase.NewLeases(deviceIP,deviceRange)
	if err:=s.DeviceLeases.ReadBindingFromDB(s.DBase);err!=nil{
		s.WriteLog("ERROR\t"+err.Error())
		os.Exit(1)
	}

	//Read PXE information
	s.Pxe.ReadPxeFromDB(s.DBase)

	if err := s.StartTCPServer(); err!=nil {
		s.WriteLog("ERROR\t"+err.Error())
		os.Exit(1)
	}
}

func stopRun(cmd *cobra.Command, args []string) {
	addr:=tcpAddr+":"+strconv.Itoa(tcpPort)
	if msg,err:=server.SendCommand(addr,"STOPCLOWDER"); err==nil {
		fmt.Println(msg)
	} else {
		fmt.Println(err.Error())
	}

}

func init() {
	stopCmd.PersistentFlags().StringVarP(&tcpAddr, "addr", "a", "localhost", "IP Address of Clowder server")
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(stopCmd)
}
