package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"go.dataddo.com/pgq"
	"log/slog"
	"math/rand"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	postgresDSN := flag.String("dsn", "", "Postgres DSN to connect to. Should be in format postgresql://user:pass@host:5432/db")
	queueName := flag.String("queue", "demo_queue", "The name of the queue to publish messages to")
	messageCount := flag.Int("count", 1, "The number of messages to publish. The weight of each message will be random")
	maxWeight := flag.Int("maxWeight", 10, "The max job weight which indicates how long the message will take the consumer to process in seconds")
	flag.Parse()

	if *postgresDSN == "" {
		panic("No postgres DNS provided")
	}

	db, err := sql.Open("pgx", *postgresDSN)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	slogger.Info("Publisher will publish messages to the queue", "count", *messageCount)

	publisher := pgq.NewPublisher(db)
	for i := 0; i < *messageCount; i++ {
		publishMessage(publisher, *queueName, rand.Intn(*maxWeight), slogger)
	}

}

func publishMessage(publisher pgq.Publisher, queueName string, weight int, slogger *slog.Logger) {
	// The payload is the json object your consumer is able to parse and understand it
	// you can use some struct common to consumers/publishers for encoding/decoding the payload
	payload := fmt.Sprintf(`{"sleep":%d, "foo": {"bar": "baz"}}`, weight)

	// The metadata is set of key=value pairs which you may or may not use
	// by using metadata you can add some extra information to the message for debugging etc.
	// Good example of metadata fields are the publisher name, payload version, trace/correlation ids...
	metadata := pgq.Metadata{
		"publisher": "localhost", // can be replaced by current pod/server name or similarly
	}

	msg := pgq.NewMessage(metadata, json.RawMessage(payload))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	id, err := publisher.Publish(ctx, queueName, msg)
	if err != nil {
		panic(err.Error())
	}

	slogger.Info("Message published", "id", id)
}
