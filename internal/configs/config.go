package configs

import (
	"fmt"
	"os"
	"runtime"
	"time"

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
	dbTickerInterval    = 10 * time.Second
	dbThresholdTimeout  = "3 MINUTE"
)

const (
	procDir      = "/proc"
	durationTime = 30 * time.Second
	algorithm    = "SHA256"
	monitorOpts  = ""
	clusterName  = "local"
)

func init() {
	fsLog := pflag.NewFlagSet("log", pflag.ContinueOnError)
	fsLog.String("verbose", "info", "verbose level")
	pflag.CommandLine.AddFlagSet(fsLog)
	if err := viper.BindPFlags(fsLog); err != nil {
		fmt.Printf("error binding flags: %v", err)
		os.Exit(2)
		return
	}

	fsSum := pflag.NewFlagSet("sum", pflag.ContinueOnError)
	fsSum.String("proc-dir", procDir, "path to /proc")
	fsSum.Duration("duration-time", durationTime, "specific interval of time repeatedly for ticker")
	fsSum.Int("count-workers", runtime.NumCPU(), "number of running workers in the workerpool")
	fsSum.String("algorithm", algorithm, "hashing algorithm for hashing data")
	fsSum.String("monitoring-options", monitorOpts, "process name and process paths to monitoring, should be represented as key=value pair. e.g. nginx=/dir1,/dir2")
	fsSum.String("cluster-name", clusterName, "Name of cluster where monitor deployed, default local")
	pflag.CommandLine.AddFlagSet(fsSum)
	if err := viper.BindPFlags(fsSum); err != nil {
		fmt.Printf("error binding flags: %v", err)
		os.Exit(1)
	}

	fsDB := pflag.NewFlagSet("db", pflag.ContinueOnError)
	fsDB.String("db-host", dbHost, "DB host")
	fsDB.Int("db-port", dbPort, "DB port")
	fsDB.String("db-name", dbName, "DB name")
	fsDB.String("db-user", dbUser, "DB user name")
	fsDB.String("db-password", dbPassword, "DB user password")
	fsDB.Int("db-connection-timeout", dbConnectionTimeout, "DB connection timeout")
	fsDB.Duration("db-ticker-interval", dbTickerInterval, "specific interval of time repeatedly for ticker")
	fsDB.String("db-threshold-timeout", dbThresholdTimeout, "specific interval of time repeatedly for query in DB")
	pflag.CommandLine.AddFlagSet(fsDB)
	if err := viper.BindPFlags(fsDB); err != nil {
		fmt.Printf("error binding flags: %v", err)
		os.Exit(1)
	}
	viper.BindEnv("db-host", "DB_HOST")
	viper.BindEnv("db-port", "DB_PORT")
	viper.BindEnv("db-name", "DB_NAME")
	viper.BindEnv("db-user", "DB_USER")
	viper.BindEnv("db-password", "DB_PASSWORD")
	viper.BindEnv("db-connection-timeout", "DB_CONNECTION_TIMEOUT")
	viper.BindEnv("db-ticker-interval", "DB_TICKER_INTERVAL")
	viper.BindEnv("db-threshold-timeout", "DB_THRESHOLD_TIMEOUT")

	fsSp := pflag.NewFlagSet("splunk", pflag.ContinueOnError)
	fsSp.Bool("splunk-enabled", false, "Enable splunk alerts")
	fsSp.String("splunk-url", "", "Splunk HTTP Events Collector URL")
	fsSp.String("splunk-token", "", "Splunk HTTP Events Collector Token")
	fsSp.Bool("splunk-insecure-skip-verify", false, "Splunk HTTP Events Collector URL skip certificate verification")
	pflag.CommandLine.AddFlagSet(fsSp)
	if err := viper.BindPFlags(fsSp); err != nil {
		fmt.Printf("error binding flags: %v", err)
		os.Exit(1)
	}
	viper.BindEnv("splunk-enabled", "SPLUNK_ENABLED")
	viper.BindEnv("splunk-url", "SPLUNK_URL")
	viper.BindEnv("splunk-token", "SPLUNK_TOKEN")
	viper.BindEnv("splunk-insecure-skip-verify", "SPLUNK_INSECURE_SKIP_VERIFY")

	fsSys := pflag.NewFlagSet("syslog", pflag.ContinueOnError)
	fsSys.Bool("syslog-enabled", false, "Enable syslog alerts")
	fsSys.String("syslog-host", "localhost", "Syslog server host")
	fsSys.Int("syslog-port", 514, "Syslog server port default 514")
	fsSys.String("syslog-proto", "tcp", "Syslog communication protocol, possible options tcp/udp, default tcp")
	pflag.CommandLine.AddFlagSet(fsSys)
	if err := viper.BindPFlags(fsSys); err != nil {
		fmt.Printf("error binding flags: %v", err)
		os.Exit(1)
	}

	viper.BindEnv("pod-namespace", "POD_NAMESPACE")
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
