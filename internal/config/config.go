package config

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

var FlagsServer FlagsServ
var FlagsAgent FlagsAg

type ConfigDB struct {
	User     string
	Adddress string
	Password string
	NameDB   string
}

type FlagsServ struct {
	Address       string
	Key           string
	Restore       bool
	StoreInterval time.Duration
	StoreFIle     string
	DBURL         string
}
type FlagsAg struct {
	Address        string
	Key            string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type Param struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	Address        string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval  time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile      string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore        bool          `env:"RESTORE" envDefault:"true"`
	Key            string        `env:"KEY"`
	DBURL          string        `env:"DATABASE_DSN"`
}
type Args struct {
	DBURL          string
	Address        string
	Key            string
	PollInterval   time.Duration
	ReportInterval time.Duration
	StoreInterval  time.Duration
	StoreFile      string
	Restore        bool
}

var ArgsM Args

func LoadConfig() Param {
	var Parametrs Param
	err := env.Parse(&Parametrs)
	if err != nil {
		log.Fatal(err)
	}
	return Parametrs
}
