#!/bin/bash

cleanup() {
    echo "Shutting down apps and containers"
    kill -INT -- -$!
    wait
    sudo docker-compose down
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start Docker services
echo "Starting Docker services..."
sudo docker-compose up -d

# Wait for PostgreSQL to be healthy
echo "Waiting for PostgreSQL to be ready..."
until [ "$(sudo docker inspect -f {{.State.Health.Status}} $(sudo docker-compose ps -q db))" = "healthy" ]; do
    sleep 2
done

export DATABASE_URL="postgres://postgres:password@localhost:5432/responses"

go run cmd/loader/main.go & loader_pid=$!
go run cmd/api/main.go & api_pid=$!
cd cmd/worker || exit 1
go run . & worker_pid=$!
cd - > /dev/null || exit 1

wait "$loader_pid" "$api_pid" "$worker_pid"