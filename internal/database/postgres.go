package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type PgxDatabase interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PgxTransactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func DatabaseConnect() *pgxpool.Pool {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("cannot load .env, " + err.Error())
	}

	db, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		panic("cannot connect to database" + err.Error())
	}

	return db
}
