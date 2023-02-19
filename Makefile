ifneq (,$(wildcard ./.env))
    include .env
    export
endif
postgres:
	docker run --name postgres15-1-alpine -p 5432:5432 -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -d postgres:15.1-alpine

runpostgres:
	docker start postgres15-1-alpine

stoppostgres:
	docker stop postgres15-1-alpine
	
createuser:
	docker exec -it postgres15-1-alpine createuser --username=$(POSTGRES_USER) grocerypl

createdb:
	docker exec -it postgres15-1-alpine createdb --username=$(POSTGRES_USER) --owner=grocerypl grocery-planner

dropdb:
	docker exec -it postgres15-1-alpine dropdb grocery-planner

dropuser:
	docker exec -it postgres15-1-alpine dropuser grocerypl

migrateup:
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(HOST):5432/grocery-planner?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(HOST):5432/grocery-planner?sslmode=disable" -verbose down

# sql generate command from docker
sqlc:
	docker run --rm -v "$(CURDIR):/src" -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createuser createdb dropdb dropuser migrateup migratedown runpostgres stoppostgres sqlc test server