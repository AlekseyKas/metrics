package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env"
)

// Init variable for flags.
var FlagsServer FlagsServ
var FlagsAgent FlagsAg

// Server flags.
type FlagsServ struct {
	Address       string
	Key           string
	StoreFile     string
	PrivateKey    string
	DBURL         string
	Config        string
	Restore       bool
	StoreInterval time.Duration
}

// Agent flags.
type FlagsAg struct {
	Address        string
	Key            string
	PubKey         string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

// Parametrs enviroment for server.
type Param struct {
	Key            string        `env:"KEY"`
	DBURL          string        `env:"DATABASE_DSN"`
	Address        string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	PubKey         string        `env:"CRYPTO_KEY"`
	PrivateKey     string        `env:"CRYPTO_KEY"`
	StoreFile      string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore        bool          `env:"RESTORE" envDefault:"true"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	StoreInterval  time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
}

// Parametrs enviroment for agent.
type Args struct {
	DBURL          string
	Address        string
	Key            string
	StoreFile      string
	PubKey         string
	PrivateKey     string
	Restore        bool
	PollInterval   time.Duration
	ReportInterval time.Duration
	StoreInterval  time.Duration
}

// Variable for environment and flags
var ArgsM Args

// Terminate flags and env
func loadConfig() Param {
	var Parametrs Param
	err := env.Parse(&Parametrs)
	if err != nil {
		log.Fatal(err)
	}
	return Parametrs
}

// Terminate flags and env with default value for server
func TermEnvFlags() {
	flag.StringVar(&FlagsServer.Address, "a", "127.0.0.1:8080", "Address")
	flag.StringVar(&FlagsServer.DBURL, "d", "", "Database URL")
	flag.StringVar(&FlagsServer.StoreFile, "f", "", "File path store")
	flag.StringVar(&FlagsServer.Key, "k", "", "Secret key")
	flag.StringVar(&FlagsServer.PrivateKey, "crypto-key", "", "Private key")
	flag.StringVar(&FlagsServer.Config, "c", "", "Path configuration file")
	flag.BoolVar(&FlagsServer.Restore, "r", true, "Restore from file")
	flag.DurationVar(&FlagsServer.StoreInterval, "i", 300000000000, "Interval store file")
	flag.Parse()
	env := loadConfig()

	envPrivateKey, _ := os.LookupEnv("CRYPTO_KEY")
	if envPrivateKey == "" {
		ArgsM.PrivateKey = FlagsServer.PrivateKey
	} else {
		ArgsM.PrivateKey = env.PrivateKey
	}
	envADDR, _ := os.LookupEnv("ADDRESS")
	if envADDR == "" {
		ArgsM.Address = FlagsServer.Address
	} else {
		ArgsM.Address = env.Address

	}
	envRest, _ := os.LookupEnv("RESTORE")
	if envRest == "" {
		ArgsM.Restore = FlagsServer.Restore
	} else {
		ArgsM.Restore = env.Restore
	}
	envStoreint, _ := os.LookupEnv("STORE_INTERVAL")
	if envStoreint == "" {
		ArgsM.StoreInterval = FlagsServer.StoreInterval
	} else {
		ArgsM.StoreInterval = env.StoreInterval
	}
	envKey, _ := os.LookupEnv("KEY")
	if envKey == "" {
		ArgsM.Key = FlagsServer.Key
	} else {
		ArgsM.Key = env.Key
	}

	envFile, b := os.LookupEnv("STORE_FILE")

	switch envFile == "" && b {
	case true:
		ArgsM.StoreFile = ""
	case false:
		if envFile == "" {
			ArgsM.StoreFile = FlagsServer.StoreFile
		} else {
			ArgsM.StoreFile = env.StoreFile
		}
	}

	envDBURL, _ := os.LookupEnv("DATABASE_DSN")

	if envDBURL == "" && FlagsServer.DBURL == "" {
		ArgsM.DBURL = ""
	} else {
		if envDBURL != "" {
			ArgsM.DBURL = envDBURL
		} else {
			ArgsM.DBURL = FlagsServer.DBURL
		}
	}
}

// Terminate flags and env with default value for agent
func TermEnvFlagsAgent() {
	flag.StringVar(&FlagsAgent.Address, "a", "127.0.0.1:8080", "Address")
	flag.StringVar(&FlagsAgent.Key, "k", "", "Secret key")
	flag.DurationVar(&FlagsAgent.ReportInterval, "r", 10000000000, "Report interval")
	flag.DurationVar(&FlagsAgent.PollInterval, "p", 2000000000, "Poll interval")
	flag.StringVar(&FlagsAgent.PubKey, "crypto-key", "", "Public key")
	flag.Parse()

	env := loadConfig()

	envPubKey, _ := os.LookupEnv("CRYPTO_KEY")
	if envPubKey == "" {
		ArgsM.PubKey = FlagsAgent.PubKey
	} else {
		ArgsM.PubKey = env.PubKey
	}

	envADDR, _ := os.LookupEnv("ADDRESS")
	if envADDR == "" {
		ArgsM.Address = FlagsAgent.Address
	} else {
		ArgsM.Address = env.Address
	}
	envRest, _ := os.LookupEnv("REPORT_INTERVAL")
	if envRest == "" {
		ArgsM.ReportInterval = FlagsAgent.ReportInterval
	} else {
		ArgsM.ReportInterval = env.ReportInterval
	}
	envStoreint, _ := os.LookupEnv("POLL_INTERVAL")
	if envStoreint == "" {
		ArgsM.PollInterval = FlagsAgent.PollInterval
	} else {
		ArgsM.PollInterval = env.PollInterval
	}
	envKey, _ := os.LookupEnv("KEY")
	if envKey == "" {
		ArgsM.Key = FlagsAgent.Key
	} else {
		ArgsM.Key = env.Key
	}
}
