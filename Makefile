.PHONY: run test test-race cover fmt vet swagger check-swagger up down logs

SWAG_DIRS := cmd/api,internal/player,internal/inventory,internal/reward,internal/leaderboard,internal/server,internal/httpapi

run:
	go run ./cmd/api

test:
	go test ./...

test-race:
	go test -race ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	gofmt -w cmd internal migrations docs

vet:
	go vet ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -d $(SWAG_DIRS) -g main.go -o docs --parseInternal

check-swagger: swagger
	git diff --exit-code -- docs

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f api activity-consumer
