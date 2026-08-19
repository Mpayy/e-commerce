migrate-up:
	migrate -path monolith/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable&x-multi-statement=true" up

migrate-down:
	migrate -path monolith/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable&x-multi-statement=true" down

mock:
	go generate ./...

product-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/product/v1/product.
	   
swag-product:
	cd services/product-service && swag init \
		-g cmd/main.go \
		-d ./,../../pkg \
		--output docs \
		--parseDependency \
		--parseInternal

swag-monolith:
	cd monolith && swag init \
		-g cmd/api/main.go \
		-d ./,../pkg \
		--output docs \
		--parseDependency \
		--parseInternal

test-unit:
	go test -cover -v ./...

test-integration:
	go test -v -race -tags=integration ./services/product-service/internal/product/repository/... 

test-all: test-unit test-integration