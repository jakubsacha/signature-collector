# Makefile

.PHONY: run-dev seed reset migrate check-db install-tools clean-tmp test

clean-tmp:
	@echo "Cleaning tmp directory..."
	@rm -rf ./tmp
	@mkdir -p ./tmp

install-tools:
	@echo "Installing development tools..."
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/cosmtrek/air@latest; }
	@command -v templ >/dev/null 2>&1 || { echo "Installing templ..."; go install github.com/a-h/templ/cmd/templ@latest; }

run-dev: clean-tmp install-tools
	@echo "Running program in development mode with hot reload..."
	@if [ ! -f .env ]; then \
		echo "Creating default .env file..."; \
		echo "LANGUAGE=pl" > .env; \
	fi
	air

seed:
	@echo "Seeding the database..."
	go run scripts/seed/main.go

migrate:
	@echo "Running migrations..."
	go run scripts/migrate/main.go

check-db:
	@echo "Checking database contents..."
	@echo ".headers on\n.mode column\nSELECT id, device_id, status, signer_name FROM documents;" | sqlite3 local.db

reset:
	@echo "Resetting database..."
	rm -f local.db
	@echo "Creating the database..."
	touch local.db
	@echo "Running migrations..."
	make migrate
	@echo "Seeding the database..."
	make seed
	@echo "\nDatabase contents:"
	@make check-db


test:
	@echo "Running tests..."
	go test -v ./...
