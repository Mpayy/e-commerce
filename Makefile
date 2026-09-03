migrate-up-monolith:
	migrate -path monolith/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce_db?sslmode=disable&x-multi-statement=true" up

migrate-down-monolith:
	migrate -path monolith/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce_db?sslmode=disable&x-multi-statement=true" down

migrate-up-notification:
	migrate -path services/notification-consumer/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/notification_db?sslmode=disable&x-multi-statement=true" up

migrate-down-notification:
	migrate -path services/notification-consumer/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/notification_db?sslmode=disable&x-multi-statement=true" down

migrate-up-user:
	migrate -path services/user-service/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/user_db?sslmode=disable&x-multi-statement=true" up

migrate-down-user:
	migrate -path services/user-service/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/user_db?sslmode=disable&x-multi-statement=true" down

migrate-up-order:
	migrate -path services/order-service/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/order_db?sslmode=disable&x-multi-statement=true" up

migrate-down-order:
	migrate -path services/order-service/database/migration -database "postgres://postgres:postgres@127.0.0.1:5432/order_db?sslmode=disable&x-multi-statement=true" down

seeder:
	docker compose exec user-service ./main --seed

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

swag-user:
	cd services/user-service && swag init \
		-g cmd/main.go \
		-d ./,../../pkg \
		--output docs \
		--parseDependency \
		--parseInternal

swag-order:
	cd services/order-service && swag init \
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
	go clean -testcache && go test -cover -v -race ./...

test-integration:
	go test -v -race -tags=integration ./services/product-service/internal/product/repository/... 

test-all: test-unit test-integration

postgres-docker:
	docker exec -it ecommerce-postgres psql -U postgres

mongo-docker:
	docker exec -it ecommerce-mongo mongosh

redis-docker:
	docker exec -it ecommerce-redis redis-cli

sqlc-order:
	cd services/order-service && sqlc generate

build-and-run:
	docker compose up -d --build

build:
	docker compose build

run:
	docker compose up -d

log-real-time:
	docker compose logs -f

status:
	docker compose ps