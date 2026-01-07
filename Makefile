# Variables
DOCKER_COMPOSE = docker compose 
DOCKER = docker 
PROJECT_NAME = melot
ENV_FILE = .env 

# Commands
help: ## Show available commands
	@echo ========================================
	@echo MELOT PROJECT - DOCKER MANAGEMENT
	@echo ========================================
	@echo LAUNCH:
	@echo   make up               - Start all services (detached)
	@echo   make down             - Stop and remove all services
	@echo   make restart          - Restart all services
	@echo   make logs             - Show all logs (follow)
	@echo ---------------------------------------
	@echo MONITORING:
	@echo   make logs-app         - Show backend logs
	@echo   make logs-bot         - Show telegram bot logs
	@echo   make logs-frontend    - Show frontend logs
	@echo   make logs-postgres    - Show database logs
	@echo   make ps               - Show container status
	@echo ---------------------------------------
	@echo DEVELOPMENT:
	@echo   make build            - Rebuild all images
	@echo   make rebuild          - Full rebuild and restart
	@echo   make shell            - Enter app container
	@echo ---------------------------------------
	@echo CLEANUP:
	@echo   make clean            - Stop and clean volumes
	@echo   make clean-images     - Remove all images
	@echo   make clean-all        - Full system cleanup
	@echo ========================================
	@echo @echo DEPENDENCIES:
	@echo ========================================
	@echo   make deps-backend    - Install Go dependencies (app + telegram)
	@echo   make deps-frontend   - Install Node.js dependencies
	@echo   make deps-all        - Install all dependencies
	@echo ---------------------------------------
	@echo FRONTEND:
	@echo   make frontend-dev    - Start frontend dev server
	@echo   make frontend-build  - Build frontend for production
	@echo   make frontend-preview- Preview production build
	@echo ========================================

# Dependencies
.PHONY: deps-backend
deps-backend: ## Install Go dependencies for backend services
	@echo "Installing Go dependencies for backend/app..."
	cd backend/app && go mod download
	@echo "Installing Go dependencies for backend/telegram..."
	cd backend/telegram && go mod download
	@echo "✓ Go dependencies installed"

.PHONY: deps-frontend
deps-frontend: ## Install Node.js dependencies
	@echo "Installing Node.js dependencies..."
	cd frontend/react-vite && npm install
	@echo "✓ Node.js dependencies installed"

.PHONY: deps-all
deps-all: deps-backend deps-frontend ## Install all dependencies

# Frontend specific
.PHONY: frontend-dev
frontend-dev: ## Start frontend development server
	@echo "Starting frontend dev server on http://localhost:5173"
	cd frontend/react-vite && npm run dev

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	@echo "Building frontend for production..."
	cd frontend/react-vite && npm run build
	@echo "✓ Frontend built to 'dist/' directory"

.PHONY: frontend-preview
frontend-preview: ## Preview production build locally
	@echo "Previewing production build..."
	cd frontend/react-vite && npm run preview

# Quick start (for local development without Docker)
.PHONY: quick-start
quick-start: deps-all ## Quick start for local development
	@echo "========================================"
	@echo "QUICK START - LOCAL DEVELOPMENT"
	@echo "========================================"
	@echo "1. Starting backend on http://localhost:8090"
	@echo "2. Starting frontend on http://localhost:5173"
	@echo "========================================"
	@echo "Open two terminals and run:"
	@echo "  Terminal 1: make backend-dev"
	@echo "  Terminal 2: make frontend-dev"
	@echo "========================================"

.PHONY: backend-dev
backend-dev: ## Start backend in development mode
	@echo "Starting backend in development mode..."
	cd backend/app && go run .


# Docker
.PHONY: up
up: ## Detouched run all services  
	$(DOCKER_COMPOSE) up -d

.PHONY: down
down: ## Stop and delete all services
	$(DOCKER_COMPOSE) down

.PHONY: restart
restart: ## Restart all services
	down up 

.PHONY: logs
logs: ## Show all logs
	$(DOCKER_COMPOSE) logs -f


.PHONY: logs-app
logs-app: ## Show all logs
	$(DOCKER_COMPOSE) logs -f app


.PHONY: logs-bot
logs-bot: ## Show all logs
	$(DOCKER_COMPOSE) logs -f telegram


.PHONY: logs-frontend
logs-frontend: ## Show all logs
	$(DOCKER_COMPOSE) logs -f frontend

.PHONY: logs-postgres
logs-postgres: ## Show all logs
	$(DOCKER_COMPOSE) logs -f postgres

# Development

.PHONY: build
build:  ## Rebuild all images
	$(DOCKER_COMPOSE) build --no-cache

.PHONY: rebuild
rebuild: down build up  ## Rebuild and start all images

.PHONY: shell
shell:  ## Check on image app (bash)
	$(DOCKER_COMPOSE) exec app bash

# Cleaning

.PHONY: clean
clean:  ## Stop all and delete volumes
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f

.PHONY: clean-images
clean-images: ## Clean all images
	docker rmi -f $$(docker images -q)

.PHONY: clean-all
clean-all: clean  ## Clear all (images, volumes, cash)
	$(DOCKER) system prune -af
	$(DOCKER) volume prune -f

.PHONY: ps
ps:  ## Show status of contnrs 
	$(DOCKER_COMPOSE) ps

.PHONY: status
status: ps  ## Show status (alias for ps)
