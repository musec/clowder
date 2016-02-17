package cmd

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/musec/clowder/dbase"
	"github.com/musec/clowder/server"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
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

func runRun(cmd *cobra.Command, args []string) {
	//Create server
	fmt.Println("Starting Clowder...")
	s, err := server.New(config)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}

	//Create log file
	logFile := config.GetString("server.log")

	var file *os.File
	if logFile == "" {
		file = os.Stdout
	} else {
		file, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Failed to open log file", logFile, ":", err)
			file = os.Stdout
		}
	}
	s.Logger = log.New(file, "", log.Ldate|log.Ltime)

	//Open databse
	dbType := config.GetString("server.dbtype")
	dbFile := config.GetString("server.database")

	s.DBase, err = dbase.Connect(dbType, dbFile)
	if err != nil {
		s.WriteLog("ERROR\t" + err.Error())
		os.Exit(1)
	}

	//Setup machine IP pool
	machineIP := net.ParseIP(config.GetString("machines.ipstart"))
	machineRange := config.GetInt("machines.iprange")
	s.MachineLeases = dbase.NewLeases(machineIP, machineRange)
	if err := s.MachineLeases.ReadBindingFromDB(s.DBase); err != nil {
		s.WriteLog("ERROR\t" + err.Error())
		os.Exit(1)
	}

	//Setup device IP pool
	deviceIP := net.ParseIP(config.GetString("devices.ipstart"))
	deviceRange := config.GetInt("devices.iprange")
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
	c := server.ConnectOrDie(config)

	if msg, err := c.SendCommand("STOPCLOWDER"); err == nil {
		fmt.Println(msg)
	} else {
		fmt.Println(err.Error())
	}

}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(stopCmd)
}
