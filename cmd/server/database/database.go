package database

import (
	"context"
	"fmt"
	"os"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

var Conn *pgxpool.Pool
var err error

//Connect to DB
func DbConnect() error {
	conf := config.ConfigDb{
		Adddress: config.ArgsM.DbURL,
		User:     "user",
		Password: "user",
		NameDb:   "db",
	}

	dbURL := "postgres://" + conf.User + ":" + conf.Password + "@" + config.ArgsM.DbURL + "/db"
	Conn, err = pgxpool.Connect(context.Background(), dbURL)

	if err != nil {
		logrus.Error("Error connection to DB: ", err)
		os.Exit(1)
	} else {
		fmt.Printf("Connected to the DB: true [" + os.Getenv("DATABASE_URL") + "] \n")
	}
	fmt.Println("44444", Conn.Ping(context.Background()))
	// Conn.Close()
	return nil
}
