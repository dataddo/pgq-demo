.PHONY: postgres setup publish consume-simple consume monitor

help: ## Print help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

POSTGRES_DSN := "postgresql://pgq:pgq@localhost:5432/postgres?sslmode=disable"
QUEUE_NAME := "demo_queue"

postgres: ## Starts the postgres 16 instance in docker container on port 5432
	docker run --name pgq-postgres -e POSTGRES_USER=pgq -e POSTGRES_PASSWORD=pgq -p 5432:5432 -d postgres:16.0

setup: ## Creates the queue table in the database
	cd setup && \
	go get && \
	go build . && \
	./setup -dsn $(POSTGRES_DSN) -queue $(QUEUE_NAME)

publish: ## Runs the publisher who publishes N messages to the queue and exits
	cd publisher && \
	go build . && \
	./publisher -dsn $(POSTGRES_DSN) -count 50 -maxWeight 10 -queue $(QUEUE_NAME)

consume-simple: ## Runs the consumer who consumes messages from the queue one by one
	cd consumer/simple/ && \
	go build . && \
	./consumer -dsn $(POSTGRES_DSN) -queue $(QUEUE_NAME)

consume-advanced: ## Runs the consumer who can consume multiple messages simultaneously
	cd consumer/advanced/ && \
	go build . && \
	./consumer -dsn $(POSTGRES_DSN) -queue $(QUEUE_NAME)

monitor: ## Runs the simple monitoring queue which collects metrics and saves it to the csv file
	cd monitor && \
	rm -f monitoring.csv && \
	go build . && \
	./monitor -dsn $(POSTGRES_DSN) -file "monitoring.csv" -queue $(QUEUE_NAME)