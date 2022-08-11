package database

import (
	"os"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

// Init database connection
var Conn *pgx.Conn

// Connect to database via DBURL
func DBConnect(DBURL string) error {
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

// Close connection to database
func DBClose() error {
	err := Conn.Close()
	if err != nil {
		logrus.Error(err)
	}
	return err
}
