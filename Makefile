.PHONY: run build test clean docker-init test-quality

# Run the application
run:
	go run main.go

# Build the application
build:
	go build -o bin/admira-etl main.go


# Initialize with docker init (run this once)
docker-init:
	docker init

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run with environment variables
dev:
	cp .env.example .env
	go run main.go

# Test individual endpoints
test-health:
	curl http://localhost:8080/healthz

test-ingest:
	curl -X POST http://localhost:8080/ingest/run

test-metrics:
	curl "http://localhost:8080/metrics/channel?from=2025-08-01&to=2025-08-10"

# NEW: Test data quality endpoints
test-quality:
	curl http://localhost:8080/quality/report

test-quality-after-ingest:
	curl -X POST http://localhost:8080/ingest/run && sleep 2 && curl http://localhost:8080/quality/report
