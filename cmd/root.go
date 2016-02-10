package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var tcpPort int

var RootCmd = &cobra.Command{Use: "clowder"}

func init() {
	var flags = RootCmd.PersistentFlags()

	flags.StringP("config", "c", "", "Configuration file")
	flags.StringP("database", "d", "/var/db/clowder.db", "Machine database")
	flags.StringP("dbtype", "t", "sqlite3", "Database type (default: sqlite3)")
	flags.IntVarP(&tcpPort, "port", "p", 5000, "TCP control port")

	viper.BindPFlag("server.database", flags.Lookup("database"))
	viper.BindPFlag("server.dbtype", flags.Lookup("dbtype"))

	err := readConfigurationFile()
	if err != nil {
		fmt.Println("Unable to open configuration file: ", err)
		os.Exit(1)
	}
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
