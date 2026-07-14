.PHONY: run build test tidy swagger seed migrate lint

## ─── Dev ────────────────────────────────────
run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server/main.go

## ─── Module ─────────────────────────────────
tidy:
	go mod tidy

## ─── Test ───────────────────────────────────
test:
	go test ./... -v -cover

## ─── Docs ───────────────────────────────────
swagger:
	swag init -g cmd/server/main.go -o docs/swagger

## ─── DB ─────────────────────────────────────
migrate:
	go run ./cmd/migrate/main.go

seed:
	go run ./cmd/seed/main.go

## ─── Lint ───────────────────────────────────
lint:
	golangci-lint run ./...

## ─── Docker ─────────────────────────────────
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f backend
