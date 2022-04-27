package database

import (
	"context"
	"fmt"
	"os"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

var Conn *pgx.Conn

//Connect to DB
func DBConnect() error {
	DBURL := config.ArgsM.DBURL
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
		fmt.Printf("Connected to the DB: true [" + os.Getenv("DATABASE_URL") + "] \n")
	}
	fmt.Println("44444", Conn.Ping(context.Background()))
	return nil
}
func DBClose() {
	Conn.Close()
}
