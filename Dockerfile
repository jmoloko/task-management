FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/app

FROM alpine:latest

WORKDIR /app

RUN mkdir -p /app/docs

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs/* ./docs/

RUN apk add --no-cache ca-certificates postgresql-client

COPY <<EOF /app/entrypoint.sh
#!/bin/sh
# Wait for PostgreSQL to be ready
until PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -U \$DB_USER -d \$DB_NAME -c "SELECT 1" > /dev/null 2>&1; do
  echo "Postgres is unavailable - sleeping"
  sleep 1
done

echo "Postgres is up - executing migrations"
PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -U \$DB_USER -d \$DB_NAME -f /app/migrations/001_init.sql

echo "Starting application"
exec ./server
EOF

RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

CMD ["/app/entrypoint.sh"]
