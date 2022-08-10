package database

import (
	"os"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

var Conn *pgx.Conn

//Connect to DB
func DBConnect(DBURL string) error {
	// DBURL := config.ArgsM.DBURL
	cfgURL, err := pgx.ParseConnectionString(DBURL)
	if err != nil {
		logrus.Error("Error parsing URL: ", err)
		return err
	}
	Conn, err = pgx.Connect(cfgURL)

	if err != nil {
		logrus.Error("Error connection to DB: ", err)
		os.Exit(1)
	} else {
		logrus.Info("Connected to the DB: true [" + os.Getenv("DATABASE_DSN") + "] \n")
	}
	return nil
}
func DBClose() error {
	err := Conn.Close()
	if err != nil {
		logrus.Error(err)
	}
	return err
}
