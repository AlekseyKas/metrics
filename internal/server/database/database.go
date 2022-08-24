package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Connect to database via DBURL
func Connect(ctx context.Context, logger *zap.Logger, DBURL string) (Conn *pgxpool.Pool, err error) {
	cfgURL, err := pgxpool.ParseConfig(DBURL)
	if err != nil {
		logger.Error("Error parsing URL: ", zap.Error(err))
		panic(err)
	}
	Conn, err = pgxpool.ConnectConfig(ctx, cfgURL)
	if err != nil {
		logger.Panic("Error connect ot database: ", zap.Error(err))
		panic(err)
	}
	return Conn, err
}
