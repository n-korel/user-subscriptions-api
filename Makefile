include .env
MIGRATIONS_PATH = ./migrations

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DSN) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DSN) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: run
run:
	@go run cmd/server/main.go

.PHONY: docker-up
docker-up:
	@docker-compose up -d

.PHONY: docker-down
docker-down:
	@docker-compose down

.PHONY: docker-logs
docker-logs:
	@docker-compose logs -f api

.PHONY: docker-build
docker-build:
	@docker-compose build --no-cache

.PHONY: gen-docs
gen-docs:
	@swag init -g ./server/main.go -d cmd,internal && swag fmt