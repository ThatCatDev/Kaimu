gql:
	go run github.com/99designs/gqlgen generate

generate: mocks gql


mocks:
	echo "Generating mocks"

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)
