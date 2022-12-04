FROM golang:latest as builder

WORKDIR /app

RUN export GO111MODULE=on

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download 
COPY . .

RUN CGO_ENABLED=0 go build -v -o ./dormitory_room_bot ./cmd/dormitory_room_bot

FROM alpine:latest

COPY --from=builder /app/dormitory_room_bot .
RUN mkdir -p ./internal/config
COPY --from=builder /app/internal/config/config.yaml ./internal/config/config.yaml

EXPOSE 8080 9000

ENTRYPOINT ["./dormitory_room_bot"]
