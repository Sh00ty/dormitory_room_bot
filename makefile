PHONY: lint
lint:
	golangci-lint run --config=.golangci.pipeline.yaml ./...

PHONY: test
test:
	go test ./...

PHONY: run
run: 
	go run ./cmd/dormitory_room_bot/main.go

PHONY: goose-up
goose-up: 
	goose -dir ./migrations postgres "user=postgres dbname=postgres password=qwerty host=127.0.0.1  port=5432" up

PHONY: goose-reset
goose-reset:
	goose -dir ./migrations postgres "user=postgres dbname=postgres password=qwerty host=127.0.0.1 port=5432" reset
