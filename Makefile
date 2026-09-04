## run: Start the server (migrations run automatically on startup)
run:
	go run cmd/api/main.go serve

## test: Run all tests
test:
	go test ./... -v

## db-up: Start PostgreSQL container
db-up:
	docker-compose up -d

## db-down: Stop PostgreSQL container
db-down:
	docker-compose down

## migrate-up: Run all pending migrations
migrate-up:
	go run cmd/api/main.go migrate up

## migrate-down: Rollback the last migration
migrate-down:
	go run cmd/api/main.go migrate down

## migrate-down-all: Rollback all migrations
migrate-down-all:
	go run cmd/api/main.go migrate down-all

## migrate-create: Create a new migration file (usage: make migrate-create name=create_profiles_table)
migrate-create:
	go run cmd/api/main.go migrate create $(name)
