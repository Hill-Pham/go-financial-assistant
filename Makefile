.PHONY: run build test lint compose-up compose-down compose-logs db-shell migrate migrate-down migrate-status

run:
	docker compose up postgres redis -d
	go run ./cmd/main.go

build:
	go build -o finassist ./cmd/main.go

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

compose-up:
	docker compose up -d --build
	docker compose logs -f app

compose-down:
	docker compose down

compose-reset:
	docker compose down -v

compose-logs:
	docker compose logs -f

logs:
	docker compose logs -f app

logs-evolution:
	docker compose logs -f evolution

db-shell:
	docker compose exec postgres psql -U finassist -d finassist
deps:
	go mod tidy
	go mod download

migrate:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

help:
	@echo ""
	@echo "Comandos disponíveis:"
	@echo "  make run              Sobe infra local + roda app Go direto"
	@echo "  make build            Compila o binário"
	@echo "  make test             Roda os testes"
	@echo "  make test-coverage    Testes com relatório de cobertura"
	@echo "  make lint             Verifica estilo e bugs"
	@echo "  make compose-up       Sobe tudo no Docker"
	@echo "  make compose-down     Derruba os containers"
	@echo "  make compose-reset    Derruba tudo + apaga volumes"
	@echo "  make compose-logs     Logs de todos os serviços"
	@echo "  make logs             Logs apenas da aplicação"
	@echo "  make logs-evolution   Logs apenas da Evolution API"
	@echo "  make db-shell         Abre o psql no container"
	@echo "  make deps             Baixa e organiza dependências"
	@echo "  make migrate          Executa todas as migrations pendentes"
	@echo "  make migrate-down     Desfaz a última migration"
	@echo "  make migrate-status   Mostra o status das migrations"
	@echo ""