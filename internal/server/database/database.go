package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

// Init database connection
// var Conn *pgxpool.Pool

// Connect to database via DBURL
func Connect(ctx context.Context, loger logrus.FieldLogger, DBURL string) (Conn *pgxpool.Pool, err error) {
	cfgURL, err := pgxpool.ParseConfig(DBURL)
	if err != nil {
		logrus.Error("Error parsing URL: ", err)
		panic(err)
	}
	Conn, err = pgxpool.ConnectConfig(ctx, cfgURL)
	if err != nil {
		logrus.Panic("Error connect ot database: ", err)
		panic(err)
	}
	return Conn, err
}

// // Close connection to database
// func DBClose(Conn *pgxpool.Pool) {
// 	Conn.Close()
// 	// if err != nil {
// 	// 	logrus.Error(err)
// 	// }
// 	return
// }
