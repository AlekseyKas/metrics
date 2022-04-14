package config

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

type Param struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	Address        string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval  time.Duration `env:"STORE_INTERVAL" envDefault:"1s"`
	StoreFile      string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore        bool          `env:"RESTORE" envDefault:"true"`
}

func LoadConfig() Param {
	var Parametrs Param
	err := env.Parse(&Parametrs)
	if err != nil {
		log.Fatal(err)
	}
	return Parametrs
}
