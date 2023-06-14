# build stage
FROM golang:1.20.5-alpine3.18 AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -x -o main ./cmd/main.go

# run stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/app.yml .
COPY --from=builder /app/.env .env
COPY --from=builder /app/resources/migration_*.sql ./resources/

EXPOSE 5000
CMD ["/app/main"]