.PHONY: setup run stop logs test vet build lint compose-config k8s-apply k8s-delete docs

COMPOSE = docker compose -f deployments/docker-compose.yml

setup:
	@if not exist .env copy .env.example .env

run: setup
	$(COMPOSE) up -d --build

stop:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f app

compose-config:
	$(COMPOSE) config

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o logpulse.exe ./cmd/api

lint:
	golangci-lint run ./...

k8s-apply:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/config-secret.yaml
	kubectl apply -f k8s/backends.yaml
	kubectl apply -f k8s/app.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/monitoring.yaml

k8s-delete:
	kubectl delete -f k8s/monitoring.yaml --ignore-not-found
	kubectl delete -f k8s/hpa.yaml --ignore-not-found
	kubectl delete -f k8s/app.yaml --ignore-not-found
	kubectl delete -f k8s/backends.yaml --ignore-not-found
	kubectl delete -f k8s/config-secret.yaml --ignore-not-found

docs:
	npx --yes http-server docs -p 8000
