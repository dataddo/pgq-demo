.PHONY: postgres setup publish consume-simple consume monitor

POSTGRES_DSN := "postgresql://pgq:pgq@localhost:5432/postgres?sslmode=disable"
QUEUE_NAME := "demo_queue"

postgres:
	docker run --name pgq-postgres -e POSTGRES_USER=pgq -e POSTGRES_PASSWORD=pgq -p 5432:5432 -d postgres:16.0

setup:
	cd setup && \
	go get && \
	go build . && \
	./setup -dsn $(POSTGRES_DSN) -queue $(QUEUE_NAME)

publish:
	cd publisher && \
	go build . && \
	./publisher -dsn $(POSTGRES_DSN) -count 50 -maxWeight 10 -queue $(QUEUE_NAME)

consume-simple:
	cd consumer/simple/ && \
	go build . && \
	./consumer -dsn $(POSTGRES_DSN) -queue $(QUEUE_NAME)

consume-advanced:
	cd consumer/advanced/ && \
	go build . && \
	./consumer -dsn $(POSTGRES_DSN) -queue $(QUEUE_NAME)

monitor:
	cd monitor && \
	rm -f monitoring.csv && \
	go build . && \
	./monitor -dsn $(POSTGRES_DSN) -file "monitoring.csv" -queue $(QUEUE_NAME)