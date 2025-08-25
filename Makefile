.PHONY: up down logs ps publish stats recent restart

COMPOSE=deploy/docker-compose.yml

up:
	docker compose -f $(COMPOSE) up --build -d

down:
	docker compose -f $(COMPOSE) down -v

logs:
	docker compose -f $(COMPOSE) logs -f --tail=200

ps:
	docker compose -f $(COMPOSE) ps

restart:
	docker compose -f $(COMPOSE) restart

publish:
	curl -s -X POST http://localhost:8080/events \
	  -H "Content-Type: application/json" \
	  -d '{"type":"signup","payload":{"user":"alice"}}' | cat

stats:
	curl -s http://localhost:8082/stats | jq .

recent:
	curl -s http://localhost:8082/events | jq .
