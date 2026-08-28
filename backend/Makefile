ENV_FILE := .env

.PHONY: dev
dev:
	@[ -f .env ] || cp .env.example .env
	docker compose up -d
	go tool air

.PHONY: down
down:
	docker compose down

.PHONY: migrate-%
migrate-%:
	go run . --env-file $(ENV_FILE) migrate $* $(ARGS)

migrate-up:
migrate-down:
migrate-reset:
migrate-create:

.PHONY: test
test:
	go test ./...
