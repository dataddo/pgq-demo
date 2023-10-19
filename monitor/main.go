package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log/slog"
	"os"
	"time"
)

func main() {
	postgresDNS := flag.String("dsn", "", "Postgres DSN to connect to. Should be in format postgresql://user:pass@host:5432/db")
	queueName := flag.String("queue", "demo_queue", "The name of the queue to monitor")
	fileName := flag.String("file", "monitoring.csv", "The name of the file where to put the monitoring metrics")
	flag.Parse()

	slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if *postgresDNS == "" {
		panic("No postgres DNS provided")
	}

	db, err := sql.Open("pgx", *postgresDNS)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	file, err := os.OpenFile(*fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	writer := csv.NewWriter(file)

	waitingQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (locked_until IS NULL OR locked_until < CURRENT_TIMESTAMP) AND processed_at IS NULL AND consumed_count < 3", *queueName)
	processingQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE locked_until is not null AND processed_at is null", *queueName)

	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				waiting := getCount(db, waitingQuery)
				processing := getCount(db, processingQuery)

				outputLine(writer, waiting, processing, t)
				slogger.Info("Monitoring", "waiting", waiting, "processing", processing)
			}
		}
	}()

	time.Sleep(10 * time.Minute)

	done <- true
}

func getCount(db *sql.DB, query string) int {
	var count int

	rows, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			panic(err.Error())
		}
	}

	return count
}

func outputLine(writer *csv.Writer, waiting int, processing int, t time.Time) {
	line := []string{
		t.Format(time.RFC3339),
		fmt.Sprintf("%d", waiting),
		fmt.Sprintf("%d", processing),
	}
	err := writer.Write(line)
	if err != nil {
		panic(err.Error())
	}

	writer.Flush()
}
