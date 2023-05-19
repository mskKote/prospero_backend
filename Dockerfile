# build stage
FROM golang:1.20.4-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/main.go

# run stage
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/app.yml .
COPY --from=builder /app/.env .env
COPY --from=builder /app/resources/migration_*.sql ./resources/

EXPOSE 5000
CMD ["/app/main"]