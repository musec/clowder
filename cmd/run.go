/*
Copyright 2015 Nhac Nguyen

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/musec/clowder/server"
	"github.com/spf13/cobra"
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

	err = s.LoadPersistentData(config)
	if err != nil {
		s.FatalError(err)
	}

	if err := s.StartTCPServer(); err != nil {
		s.FatalError(err)
		os.Exit(1)
	}
}

func stopRun(cmd *cobra.Command, args []string) {
	server.Exec("STOPCLOWDER", config)
}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(stopCmd)
}
