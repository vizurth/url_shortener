COMPOSE_FILE  := docker-compose.yml

.PHONY: up down test test-integration smoke bench lint mock

up:
	docker-compose -f $(COMPOSE_FILE) up -d --build

down:
	docker-compose -f $(COMPOSE_FILE) down -v

test:
	go test -race ./...

test-integration:
	go test -race -tags=integration -timeout=60s ./...

smoke:
	go run ./cmd/smoketest

bench:
	go test -bench=. -benchtime=30s -benchmem ./bench/...

lint:
	golangci-lint run ./...

mock:
	mockery