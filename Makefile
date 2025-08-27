.PHONY: up up-debug down logs logs50 ps rebuild

PROJECT=goopt

up:
	docker compose -p $(PROJECT) up -d --build

up-debug:
	docker compose -p $(PROJECT) -f docker-compose.yml -f docker-compose.debug.yml up -d --build

down:
	docker compose -p $(PROJECT) down -v

logs:
	docker compose -p $(PROJECT) logs -f --tail=200

logs50:
	docker compose -p $(PROJECT) logs --tail=50

ps:
	docker compose -p $(PROJECT) ps

rebuild:
	docker compose -p $(PROJECT) build --no-cache
