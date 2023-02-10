package configs

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// config defaults
const (
	dbHost              = "127.0.0.1"
	dbPort              = 5432
	dbName              = "postgres"
	dbUser              = "postgres"
	dbPassword          = "postgres"
	dbConnectionTimeout = 10
)

func init() {
	fsDB := pflag.NewFlagSet("db", pflag.ContinueOnError)
	fsDB.String("db-host", dbHost, "DB host")
	fsDB.Int("db-port", dbPort, "DB port")
	fsDB.String("db-name", dbName, "DB name")
	fsDB.String("db-user", dbUser, "DB user name")
	fsDB.String("db-password", dbPassword, "DB user password")
	fsDB.Int("db-connection-timeout", dbConnectionTimeout, "DB connection timeout")
	if err := viper.BindPFlags(fsDB); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetDBConnString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=%d",
		viper.GetString("db-user"),
		viper.GetString("db-password"),
		viper.GetString("db-host"),
		viper.GetInt("db-port"),
		viper.GetString("db-name"),
		viper.GetInt("db-connection-timeout"),
	)
}
