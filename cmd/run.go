package cmd

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-homedir"
	"github.com/musec/clowder/dbase"
	"github.com/musec/clowder/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"path"
	"strconv"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Clowder server management service.",
	Long: `Start Clowder server management service.
	`,
	Run: runRun,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Close Clowder server management service.",
	Long: `Close Clowder server management service.
	`,
	Run: stopRun,
}

func readConfigurationFile() error {
	viper.SetConfigName("clowder")

	// Prefer user configuration to local configuration
	// to distribution configuration, etc.
	homedir, err := homedir.Dir()
	if err != nil {
		return err
	}

	viper.AddConfigPath(homedir)
	viper.AddConfigPath(".")
	viper.AddConfigPath(path.Join(homedir, ".clowder"))
	viper.AddConfigPath(path.Join(homedir, "clowder"))
	viper.AddConfigPath(path.Join(homedir, ".config"))
	viper.AddConfigPath(path.Join(homedir, ".config", "clowder"))
	viper.AddConfigPath("/usr/local/etc")
	viper.AddConfigPath("/etc")

	err = viper.ReadInConfig()
	if notfound, ok := err.(*viper.ConfigFileNotFoundError); ok {
		fmt.Println(notfound, "- using default settings")
	}

	return err
}

func runRun(cmd *cobra.Command, args []string) {
	err := readConfigurationFile()

	if err != nil {
		fmt.Println("Unable to open configuration file: ", err)
		os.Exit(1)
	}

	//Create server
	fmt.Println("Starting Clowder...")
	serverIP := net.ParseIP(viper.GetString("server.ip")).To4()
	serverMask := net.ParseIP(viper.GetString("server.subnetmask")).To4()
	duration := viper.GetDuration("server.duration")
	hostname, _ := os.Hostname()
	dns := net.ParseIP(viper.GetString("server.dns")).To4()
	router := net.ParseIP(viper.GetString("server.router")).To4()
	domainName := viper.GetString("server.domainname")
	dbFile := viper.GetString("database.filename")
	s := server.NewServer(serverIP, serverMask, tcpPort, duration, hostname, dns, router, domainName)

	//Create log file
	logFile := "clowder.log"
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to open log file", logFile, ":", err)
		file = os.Stdout
	}
	s.Logger = log.New(file, "", log.Ldate|log.Ltime)

	//Open databse
	if db, err := sql.Open("sqlite3", dbFile); err == nil && db != nil {
		s.DBase = db
	} else {
		s.WriteLog("ERROR\t" + err.Error())
		os.Exit(1)
	}

	//Setup machine IP pool
	machineIP := net.ParseIP(viper.GetString("machines.ipstart"))
	machineRange := viper.GetInt("machines.iprange")
	s.MachineLeases = dbase.NewLeases(machineIP, machineRange)
	if err := s.MachineLeases.ReadBindingFromDB(s.DBase); err != nil {
		s.WriteLog("ERROR\t" + err.Error())
		os.Exit(1)
	}

	//Setup device IP pool
	deviceIP := net.ParseIP(viper.GetString("devices.ipstart"))
	deviceRange := viper.GetInt("devices.iprange")
	s.DeviceLeases = dbase.NewLeases(deviceIP, deviceRange)
	if err := s.DeviceLeases.ReadBindingFromDB(s.DBase); err != nil {
		s.WriteLog("ERROR\t" + err.Error())
		os.Exit(1)
	}

	//Read PXE information
	s.Pxe.ReadPxeFromDB(s.DBase)

	if err := s.StartTCPServer(); err != nil {
		s.WriteLog("ERROR\t" + err.Error())
		os.Exit(1)
	}
}

func stopRun(cmd *cobra.Command, args []string) {
	addr := tcpAddr + ":" + strconv.Itoa(tcpPort)
	if msg, err := server.SendCommand(addr, "STOPCLOWDER"); err == nil {
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
