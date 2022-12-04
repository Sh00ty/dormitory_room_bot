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
	goose -dir ./migrations postgres "user= dbname= password= host= port=" up

PHONY: goose-reset
goose-reset:
	goose -dir ./migrations postgres "user= dbname= password= host= port=" reset

