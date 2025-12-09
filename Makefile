.PHONY: run stop build test lint clean setup

# Project Name
APP_NAME = logpulse

# Setup: Auto create .env (Cross-platform compatible for sh/bash)
setup:
	@echo "Initializing .env file..."
	@if [ ! -f .env ]; then cp .env.example .env; fi

# Run: Setup first
run: setup
	docker-compose -f deployments/docker-compose.yml up -d --build
stop:
	docker-compose -f deployments/docker-compose.yml down

logs:
	docker-compose -f deployments/docker-compose.yml logs -f app


# local
test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf tmp/
	docker system prune -f