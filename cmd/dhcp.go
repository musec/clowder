/*
 * Copyright 2015 Nhac Nguyen
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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
	server.Exec("DHCPON", config)
}

func dhcpOffRun(cmd *cobra.Command, args []string) {
	server.Exec("DHCPOFF", config)
}

func statRun(cmd *cobra.Command, args []string) {
	server.Exec("STATUS", config)
}
