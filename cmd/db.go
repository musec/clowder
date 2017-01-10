/*
 * Copyright (c) 2016 Jonathan Anderson
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/musec/clowder/db"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func getDB(log *log.Logger) db.DB {
	dbtype := config.GetString("server.dbtype")
	dbname := config.GetString("server.database")

	db, err := db.Open(dbtype, dbname, log)
	if err != nil {
		fmt.Println("Error opening", dbtype, "database at", dbname,
			":", err)
		os.Exit(1)
	}

	return db
}

var dbCmd = cobra.Command{
	Use:   "db",
	Short: "Interact with the Clowder database",
}

func init() {
	RootCmd.AddCommand(&dbCmd)
}
