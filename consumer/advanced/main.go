package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go.dataddo.com/pgq"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	ctx := context.Background()

	postgresDSN := flag.String("dsn", "", "Postgres DSN to connect to. Should be in format postgresql://user:pass@host:5432/db")
	queueName := flag.String("queue", "demo_queue", "The name of the queue to consume")
	flag.Parse()

	db, err := sql.Open("pgx", *postgresDSN)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	consumer, err := newConsumer(db, *queueName)
	if err != nil {
		panic(err.Error())
	}

	err = consumer.Run(ctx)
	if err != nil {
		panic(err.Error())
	}
}

type handler struct {
	worker Worker
}

func (h *handler) HandleMessage(_ context.Context, msg pgq.Message) (processed bool, err error) {
	var job Job
	err = json.Unmarshal(msg.Payload(), &job)
	if err != nil {
		return pgq.MessageNotProcessed, err
	}

	err = h.worker.Do(job)
	if err != nil {
		return pgq.MessageProcessed, err
	}

	return pgq.MessageProcessed, nil
}

func newConsumer(db *sql.DB, queueName string) (*pgq.Consumer, error) {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	h := handler{worker: Worker{}}

	return pgq.NewConsumer(db, queueName, &h,
		pgq.WithLogger(slogger),
		pgq.WithMaxParallelMessages(5),
		pgq.WithLockDuration(2*time.Minute),
		pgq.WithPollingInterval(500*time.Millisecond),
		// add other options here if you wish, please see the docs https://github.com/dataddo/pgq#consumer-options
	)
}

type Job struct {
	Id    string `json:"id"`
	Sleep int    `json:"sleep"`
}

type Worker struct {
}

func (w *Worker) Do(job Job) error {
	slog.Info("Sleeper started to work on the job.", "job", job)
	time.Sleep(time.Duration(job.Sleep) * time.Second)

	if job.Sleep == 0 {
		slog.Error("Sleeper failed to process the job.", "job", job)
		return errors.New(fmt.Sprintf("Gophers need to sleep!. Sleeping for '%d' is unacceptable for them.", job.Sleep))
	}

	slog.Info("Sleeper finished the job.", "job", job)
	return nil
}
