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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var config *viper.Viper

var RootCmd = &cobra.Command{Use: "clowder"}

func init() {
	var flags = RootCmd.PersistentFlags()
	config = viper.New()

	flags.StringP("config", "c", "", "Configuration file")

	flags.StringP("database", "d", "/var/db/clowder.db", "Machine database")
	config.BindPFlag("server.database", flags.Lookup("database"))

	flags.StringP("dbtype", "t", "sqlite3", "Database type (default: sqlite3)")
	config.BindPFlag("server.dbtype", flags.Lookup("dbtype"))

	flags.String("host", "localhost", "Clowder server")
	config.BindPFlag("server.host", flags.Lookup("host"))

	flags.IntP("port", "p", 5000, "TCP control port")
	config.BindPFlag("server.controlPort", flags.Lookup("port"))

	err := readConfigurationFile()
	if err != nil {
		fmt.Println("Unable to open configuration file: ", err)
		os.Exit(1)
	}
}

func readConfigurationFile() error {
	config.SetConfigName("clowder")

	// Prefer user configuration to local configuration
	// to distribution configuration, etc.
	homedir, err := homedir.Dir()
	if err != nil {
		return err
	}

	config.AddConfigPath(homedir)
	config.AddConfigPath(".")
	config.AddConfigPath(path.Join(homedir, ".clowder"))
	config.AddConfigPath(path.Join(homedir, "clowder"))
	config.AddConfigPath(path.Join(homedir, ".config"))
	config.AddConfigPath(path.Join(homedir, ".config", "clowder"))
	config.AddConfigPath("/usr/local/etc")
	config.AddConfigPath("/etc")

	err = config.ReadInConfig()
	if notfound, ok := err.(*viper.ConfigFileNotFoundError); ok {
		fmt.Println(notfound, "- using default settings")
	}

	return err
}
