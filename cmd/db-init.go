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

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("Initializing database...")

	db := getDB()

	err := db.Init()
	if err != nil {
		fmt.Println("Error initializing database: ", err)
		os.Exit(1)
	}
}

func init() {
	dbCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize a new Clowder database",
		Run:   runInit,
	})
}
