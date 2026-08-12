migrate-up:
	migrate -path database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable&x-multi-statement=true" up

migrate-down:
	migrate -path database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable&x-multi-statement=true" down

mock:
	go generate ./...