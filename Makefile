.PHONY: help up down db-up db-down api-test api-fmt api-tidy api-run web-install web-dev web-build web-test

help:
	@echo "TEAmate local dev commands"
	@echo ""
	@echo "Backend (Go)"
	@echo "  make up        - start db+api via docker compose"
	@echo "  make down      - stop compose services"
	@echo "  make db-up     - start only postgres (docker compose)"
	@echo "  make db-down   - stop only postgres"
	@echo "  make api-run   - run api locally (requires DATABASE_URL)"
	@echo "  make api-test  - run api unit tests"
	@echo "  make api-fmt   - go fmt ./..."
	@echo "  make api-tidy  - go mod tidy"
	@echo ""
	@echo "Frontend (Next.js)"
	@echo "  make web-install - npm install in web/"
	@echo "  make web-dev     - run next dev"
	@echo "  make web-build   - run next build"
	@echo "  make web-test    - run web tests (currently lint)"

up:
	@docker compose -f build/docker-compose.yml up --build

down:
	@docker compose -f build/docker-compose.yml down

db-up:
	@docker compose -f build/docker-compose.yml up -d db

db-down:
	@docker compose -f build/docker-compose.yml stop db

api-test:
	@cd api && go test ./...

api-fmt:
	@cd api && go fmt ./...

api-tidy:
	@cd api && go mod tidy

api-run:
	@cd api && go run ./cmd/serverd

web-install:
	@cd web && npm install

web-dev:
	@cd web && npm run dev

web-build:
	@cd web && npm run build

web-test:
	@cd web && npm test
