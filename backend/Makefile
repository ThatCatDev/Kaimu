.PHONY: setup deps generate gql mocks migrate-create migrate-up migrate-down run

setup: deps generate
	@echo "Setup complete! Run 'make run' to start the server."

deps:
	go mod download
	go mod tidy

gql:
	go run github.com/99designs/gqlgen generate

generate: mocks gql

mocks:
	@echo "Generating mocks"

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

migrate-up:
	go run cmd/main.go migrate up

migrate-down:
	go run cmd/main.go migrate down

run:
	go run cmd/main.go serve
