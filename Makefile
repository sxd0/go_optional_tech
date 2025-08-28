.PHONY: up up-debug down logs logs50 ps rebuild gen seed grpc-stats grpc-user rest-stats diag-kafka diag-redis test-smoke test-reset

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

gen:
	protoc -I proto \
	  --go_out=proto/gen/go --go_opt=paths=source_relative \
	  --go-grpc_out=proto/gen/go --go-grpc_opt=paths=source_relative \
	  proto/stats/v1/stats.proto

seed:
	curl -s -X POST http://localhost:18080/v1/publish -H 'content-type: application/json' -d '{"user_id":1,"action":"click"}' >/dev/null
	curl -s -X POST http://localhost:18080/v1/publish -H 'content-type: application/json' -d '{"user_id":1,"action":"view"}' >/dev/null

grpc-stats:
	grpcurl -plaintext -d '{}' localhost:19090 stats.v1.StatsService/GetStats

grpc-user:
	grpcurl -plaintext -d '{"userId":1}' localhost:19090 stats.v1.StatsService/GetUserLastAction

