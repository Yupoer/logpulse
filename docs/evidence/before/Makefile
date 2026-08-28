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

# --- Testing Tools ---

.PHONY: test-load

test-load: ## sequential round-robin 3 Service Log
	@echo "Starting Load Test (Sequential Round-Robin)..."
	@bash -c 'services=("payment" "user" "order"); \
	for i in {1..12}; do \
		idx=$$(( (i - 1) % 3 )); \
		svc=$${services[$$idx]}; \
		echo " [$$i] Sending log from $$svc-service (Partition $$idx candidate)..."; \
		curl -s -X POST http://localhost/logs \
		-H "Content-Type: application/json" \
		-d "{\"service_name\": \"$$svc-service\", \"level\": \"INFO\", \"message\": \"Sequential test log $$i\"}" \
		-w "\n"; \
	done'
	@echo " Load test finished. Run 'make check-consumers' to see balanced distribution!"

.PHONY: check-consumers

check-consumers: ## check Kafka Consumer Group
	@docker exec -it logpulse-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --describe --group logpulse-group