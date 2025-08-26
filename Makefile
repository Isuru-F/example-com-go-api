APP_NAME=ecom-book-store-sample-api
BIN_DIR=bin
BIN_PATH=$(BIN_DIR)/$(APP_NAME)

.PHONY: build run clean test test-api deps

build: deps
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) ./cmd

run: deps
	go run ./cmd

clean:
	rm -rf $(BIN_DIR)
	go clean ./...

deps:
	go mod tidy

# Basic unit tests
test:
	go test ./...

# Quick manual API test using curl.
# Starts the server in the background, hits a few endpoints, then stops it.
# Requires a POSIX shell.

SHELL := /bin/bash

PORT?=8080

test-api: build
	set -euo pipefail; \
	./$(BIN_PATH) & PID=$$!; \
	sleep 1; \
	echo "List products"; \
	curl -s http://localhost:$(PORT)/api/v1/products | head -c 200; echo; \
	echo "Create product"; \
	curl -s -X POST http://localhost:$(PORT)/api/v1/products -H 'Content-Type: application/json' -d '{"title":"Test Book","author":"Tester","description":"Desc","price":19.99,"stock":10}'; echo; \
	echo "Add to cart"; \
	curl -s -X POST http://localhost:$(PORT)/api/v1/cart/user/1/items -H 'Content-Type: application/json' -d '{"productId":1,"quantity":2}'; echo; \
	echo "Place order"; \
	curl -s -X POST http://localhost:$(PORT)/api/v1/orders/user/1; echo; \
	kill $$PID || true
