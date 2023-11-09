package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.dataddo.com/pgq/x/schema"
)

func main() {
	postgresDSN := flag.String("dsn", "", "Postgres DSN to connect to. Should be in format postgresql://user:pass@host:5432/db")
	queueName := flag.String("queue", "demo_queue", "The name of the queue to setup")
	flag.Parse()

	if *postgresDSN == "" {
		panic("No postgres DNS provided")
	}

	db, err := sql.Open("pgx", *postgresDSN)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	ctx := context.Background()
	_, err = db.ExecContext(ctx, schema.GenerateDropTableQuery(*queueName))
	if err != nil {
		panic(err.Error())
	}

	_, err = db.ExecContext(ctx, schema.GenerateCreateTableQuery(*queueName))
	if err != nil {
		panic(err.Error())
	}
}
