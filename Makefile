.PHONY: up down build logs backend-logs frontend-logs db-shell redis-shell

up:
	docker compose up --build

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

backend-logs:
	docker compose logs -f backend

frontend-logs:
	docker compose logs -f frontend

db-shell:
	docker compose exec postgres psql -U faultline -d faultline

redis-shell:
	docker compose exec redis redis-cli

clean:
	docker compose down -v --remove-orphans
