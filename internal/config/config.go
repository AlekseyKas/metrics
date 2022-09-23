package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
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
	TrustedSubnet string
	GRPC          bool
	Restore       bool
	StoreInterval time.Duration
}

// Agent flags.
type FlagsAg struct {
	Address        string
	Key            string
	PubKey         string
	Config         string
	GRPC           bool
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
	Config         string        `env:"CONFIG"`
	TrustedSubnet  string        `env:"TRUSTED_SUBNET"`
	Restore        bool          `env:"RESTORE" envDefault:"true"`
	GRPC           bool          `env:"GRPC"`
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
	Config         string
	TrustedSubnet  string
	Restore        bool
	GRPC           bool
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
	flag.StringVar(&FlagsServer.Config, "config", "", "Path configuration file")
	flag.StringVar(&FlagsServer.TrustedSubnet, "t", "", "Trusted subnet")
	flag.BoolVar(&FlagsServer.Restore, "r", true, "Restore from file")
	flag.BoolVar(&FlagsServer.GRPC, "g", false, "GRPC")
	flag.DurationVar(&FlagsServer.StoreInterval, "i", 300000000000, "Interval store file")
	flag.Parse()
	env := loadConfig()

	envGRPC, _ := os.LookupEnv("GRPC")
	if envGRPC == "" {
		ArgsM.GRPC = FlagsServer.GRPC
	} else {
		ArgsM.GRPC = env.GRPC
	}

	envTrustedSubnet, _ := os.LookupEnv("TRUSTED_SUBNET")
	if envTrustedSubnet == "" {
		ArgsM.TrustedSubnet = FlagsServer.TrustedSubnet
	} else {
		ArgsM.TrustedSubnet = env.TrustedSubnet
	}

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
	envConfig, _ := os.LookupEnv("CONFIG")
	if envConfig != "" && FlagsServer.Config == "" {
		parseConfig(envConfig)
	}
	if envConfig == "" && FlagsServer.Config != "" {
		parseConfig(FlagsServer.Config)
	}
}

// Var for unmarshal duration type
type Duration time.Duration

// Parametrs enviroment for agent.
type Config struct {
	DatabaseDSN    string   `json:"database_dsn"`
	CryptoKey      string   `json:"crypto_key"`
	Address        string   `json:"address"`
	StoreFile      string   `json:"store_file"`
	TrustedSubnet  string   `json:"trusted_subnet"`
	Restore        bool     `json:"restore"`
	StoreInterval  Duration `json:"store_interval"`
	ReportInterval Duration `json:"report_interval"`
	PollInterval   Duration `json:"poll_interval"`
}

// Parse config.
func parseConfig(configPath string) error {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		logrus.Error(err)
	}
	if ArgsM.Address == "" {
		ArgsM.Address = config.Address
	}
	if ArgsM.PrivateKey == "" {
		ArgsM.PrivateKey = config.CryptoKey
	}
	if ArgsM.DBURL == "" {
		ArgsM.DBURL = config.DatabaseDSN
	}
	if ArgsM.StoreFile == "" {
		ArgsM.StoreFile = config.StoreFile
	}
	if ArgsM.TrustedSubnet == "" {
		ArgsM.TrustedSubnet = config.TrustedSubnet
	}
	return err
}

// Terminate flags and env with default value for agent
func TermEnvFlagsAgent() {
	flag.StringVar(&FlagsAgent.Address, "a", "127.0.0.1:8080", "Address")
	flag.StringVar(&FlagsAgent.Key, "k", "", "Secret key")
	flag.StringVar(&FlagsAgent.Config, "c", "", "Path configuration file")
	flag.StringVar(&FlagsAgent.Config, "config", "", "Path configuration file")
	flag.DurationVar(&FlagsAgent.ReportInterval, "r", 10000000000, "Report interval")
	flag.DurationVar(&FlagsAgent.PollInterval, "p", 2000000000, "Poll interval")
	flag.BoolVar(&FlagsAgent.GRPC, "g", false, "GRPC")
	flag.StringVar(&FlagsAgent.PubKey, "crypto-key", "", "Public key")
	flag.Parse()

	env := loadConfig()

	envGRPC, _ := os.LookupEnv("GRPC")
	if envGRPC == "" {
		ArgsM.GRPC = FlagsAgent.GRPC
	} else {
		ArgsM.GRPC = env.GRPC
	}

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
	envConfig, _ := os.LookupEnv("CONFIG")
	if envConfig != "" && FlagsAgent.Config == "" {
		parseConfig(envConfig)
	}
	if envConfig == "" && FlagsAgent.Config != "" {
		parseConfig(FlagsAgent.Config)
	}
}
