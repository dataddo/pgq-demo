package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"go.dataddo.com/pgq"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	ctx := context.Background()

	postgresDNS := flag.String("dsn", "", "Postgres DSN to connect to. Should be in format postgresql://user:pass@host:5432/db")
	queueName := flag.String("queue", "demo_queue", "The name of the queue to consume")
	flag.Parse()

	db, err := sql.Open("pgx", *postgresDNS)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	h := handler{worker: Worker{}}
	consumer, err := pgq.NewConsumer(db, *queueName, &h, pgq.WithLockDuration(3*time.Minute))
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
		return pgq.MessageNotProcessed, err
	}

	return pgq.MessageProcessed, nil
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
	slog.Info("Sleeper finished the job.", "job", job)

	return nil
}
