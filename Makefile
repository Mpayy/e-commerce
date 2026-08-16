migrate-up:
	migrate -path monolith/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable&x-multi-statement=true" up

migrate-down:
	migrate -path monolith/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable&x-multi-statement=true" down

mock:
	go generate ./...

product-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/product/v1/product.proto

test-unit:
	go test ./...

test-integration:
	go test -v -race -tags=integration ./service/product-service/internal/product/repository/... 

test-all: test-unit test-integration