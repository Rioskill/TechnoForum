package utils

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dsn = "user=postgres dbname=forum password=12345 host=localhost port=5432 sslmode=disable pool_max_conns=20"

func InitPostgres() (*pgxpool.Pool, error) {
	till := time.Now().Add(time.Second * 10)

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	for time.Now().Before(till) {
		log.Println("Trying to open pg connection")

		err = db.Ping(context.Background())
		if err == nil {
			log.Println("Ping sucessful")
			break
		}

		time.Sleep(time.Second)
	}

	return db, nil
}

func MakeTx(db *pgxpool.Pool, fb func(tx pgx.Tx) error) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}

	err = fb(tx)
	if err != nil {
		rollBackErr := tx.Rollback(context.Background())
		if rollBackErr != nil {
			return rollBackErr
		}
		return err
	}

	return tx.Commit(context.Background())
}
