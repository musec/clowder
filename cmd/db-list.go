/*
 * Copyright (c) 2015 Nhac Nguyen
 * Copyright (c) 2016 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"os"
)

func runList(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		os.Exit(1)
	}

	db := getDB()

	for _, arg := range args {
		if arg == "machines" {
			fmt.Println("Machines:")
			rows, err := db.GetMachines()

			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}

			for _, row := range rows {
				fmt.Println(row)
			}

		} else if arg == "users" {
			fmt.Println("Users:")

			rows, err := db.GetUsers()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}

			for _, row := range rows {
				fmt.Println(row)
			}

		} else if arg == "reservations" {
			fmt.Println("Reservations:")

			rows, err := db.GetReservations()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}

			for _, row := range rows {
				fmt.Println(row)
			}

		} else {
			fmt.Println("Unknown table: ", arg)
			os.Exit(1)
		}

		fmt.Println("")
	}
}

func init() {
	dbCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List data from the Clowder database",
		Run:   runList,
	})
}
