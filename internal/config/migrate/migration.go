package migrate

import (
	"context"
	"embed"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

// MigrateFromDir executes database migrations
func MigrateFromDir(ctx context.Context, db *pgx.Conn, migrationDir string, loger logrus.FieldLogger) error {
	if loger == nil {
		loger = logrus.StandardLogger()
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	createMigrationTable := `
		CREATE TABLE IF NOT EXISTS migration(
			id          varchar(255) primary key,
			modified_at timestamp not null
		);
	`
	var com pgconn.CommandTag
	if com, err = tx.Exec(ctx, createMigrationTable); err != nil {
		return err
	}
	logrus.Debug(string(com))

	if _, err = tx.Exec(ctx, `LOCK TABLE migration;`); err != nil {
		return err
	}

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			return err
		}
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, f := range files {
		fileName := f.Name()

		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		filePath := path.Join(migrationDir, fileName)

		r := tx.QueryRow(ctx, `SELECT id, modified_at FROM migration WHERE id = $1;`, fileName)

		type migrationItem struct {
			ModifiedAt time.Time
			ID         string
		}

		mi := &migrationItem{}
		err := r.Scan(&mi.ID, &mi.ModifiedAt)

		if err != nil && err != pgx.ErrNoRows {
			err = tx.Rollback(ctx)
			if err != nil {
				return err
			}
			return err
		} else if err == nil {
			continue
		}

		script, err := os.ReadFile(filePath)
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				return err
			}
			return err
		}
		loger.Info(string(script))
		if _, err := tx.Exec(ctx, string(script)); err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				return err
			}
			return err
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO migration (id, modified_at) VALUES($1, $2) ON CONFLICT (id) DO UPDATE SET modified_at = $2;`,
			fileName, time.Now().UTC(),
		); err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				return err
			}
			return err
		}
	}
	return tx.Commit(ctx)
}

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Close(context.Context) error
}

// MigrateFromFS executes database migrations
func MigrateFromFS(ctx context.Context, db *pgxpool.Pool, migrations *embed.FS, loger logrus.FieldLogger) error {
	if loger == nil {
		loger = logrus.StandardLogger()
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	createMigrationTable := `
		CREATE TABLE IF NOT EXISTS migration(
			id          varchar(255) primary key,
			modified_at timestamp not null
		);
	`
	var commandTag pgconn.CommandTag
	if commandTag, err = tx.Exec(ctx, createMigrationTable); err != nil {
		return err
	}
	logrus.Debug(commandTag)
	if commandTag, err = tx.Exec(ctx, `LOCK TABLE migration;`); err != nil {
		return err
	}
	logrus.Debug(commandTag)
	files, err := migrations.ReadDir(".")
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			loger.Error(err)
			return err
		}
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	// changed
	type migrationItem struct {
		ModifiedAt time.Time
		ID         string
	}
	for _, f := range files {
		fileName := f.Name()
		r := tx.QueryRow(ctx, `SELECT id, modified_at FROM migration WHERE id = $1;`, fileName)

		mi := &migrationItem{}
		err := r.Scan(&mi.ID, &mi.ModifiedAt)

		if err != nil && err != pgx.ErrNoRows {
			err = tx.Rollback(ctx)
			if err != nil {
				loger.Error(err)
				return err
			}
			return err
		} else if err == nil {
			continue
		}

		script, err := migrations.ReadFile(fileName)
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				loger.Error(err)
				return err
			}
			return err
		}
		loger.Info(string(script))
		if _, err := tx.Exec(ctx, string(script)); err != nil {
			loger.Error(err)
			err = tx.Rollback(ctx)
			if err != nil {
				loger.Error(err)
				return err
			}
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO migration (id, modified_at) VALUES($1, $2) ON CONFLICT (id) DO UPDATE SET modified_at = $2;`,
			fileName, time.Now().UTC(),
		); err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				loger.Error(err)
				return err
			}
			return err
		}
	}
	return tx.Commit(ctx)
}
