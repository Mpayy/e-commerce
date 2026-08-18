<div align="center">

# 🛒 E-Commerce API

### Go · Gin · GORM · PostgreSQL · MongoDB · Redis · gRPC · RabbitMQ · JWT · Docker

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12.0-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://github.com/gin-gonic/gin)
[![GORM](https://img.shields.io/badge/GORM-v1.31.2-E10098?style=for-the-badge)](https://gorm.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![MongoDB](https://img.shields.io/badge/MongoDB-8-47A248?style=for-the-badge&logo=mongodb&logoColor=white)](https://www.mongodb.com/)
[![Redis](https://img.shields.io/badge/Redis-v9-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![gRPC](https://img.shields.io/badge/gRPC-v1.83.0-244C5A?style=for-the-badge&logo=grpc&logoColor=white)](https://grpc.io/)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-4-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)](https://www.rabbitmq.com/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docs.docker.com/compose/)
[![Swagger](https://img.shields.io/badge/Swagger-OpenAPI-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)](http://localhost:8081/swagger/index.html)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=for-the-badge)]()</p>

<p align="center">
  <em>A hybrid monorepo e-commerce backend: a Go monolith (User/Cart/Order) communicating with independently deployed microservices (Product Catalog, Notification Consumer) via gRPC and RabbitMQ.<br>
  Structured on Clean Architecture, with explicit service boundaries enforced by protobuf contracts and a saga-based compensation pattern for distributed stock consistency.</em>
</p>

</div>

---

## 📋 Table of Contents

- [Features & Roadmap](#-features--roadmap)
- [Tech Stack](#-tech-stack)
- [Architecture Overview](#-architecture-overview)
- [Directory Structure](#-directory-structure)
- [API Endpoints](#-api-endpoints)
- [API Documentation](#-api-documentation)
- [Request & Response Samples](#-request--response-samples)
- [Getting Started](#-getting-started)
- [Architecture Notes](#-architecture-notes)

---

## ✅ Features & Roadmap

### 🟢 Implemented

**Authentication & User** *(Monolith — PostgreSQL)*
- [x] **User Registration** — Password hashing with bcrypt; domain-specific error on duplicate email
- [x] **JWT Authentication (HS256)** — 30-day tokens; secret key loaded from environment config via Viper
- [x] **Redis Session Management** — JWT stored in Redis (`auth:session:<token>`) on login; atomically deleted on logout for immediate invalidation
- [x] **Auth Middleware** — Validates Bearer token signature, expiry, and Redis session existence; injects `auth` context
- [x] **Role-Based Access Control (RBAC)** — Admin middleware reads role from JWT claims; returns `403` for unauthorized access
- [x] **User Profile** — Returns authenticated user data resolved from `user_id` in the JWT

**Product Catalog** *(Product Service — MongoDB, port 8081)*
- [x] **Category Management** — Admin: create categories with auto-generated slugs, unique index on slug enforced at MongoDB level
- [x] **Product — Create** — Admin: validates category, auto-generates SKU if not provided; unique indexes on slug and SKU
- [x] **Product — Update** — Admin: full field update including `is_active` toggle; re-validates category only when changed
- [x] **Product — Soft Delete** — Admin: sets `is_active = false` via atomic `$set` — no row deletion
- [x] **Product — Listing & Search** — Public: paginated list with optional `search` (regex, case-insensitive), `category_id`, `page`, and `limit`; sorted by name ASC
- [x] **Product — Detail** — Public: returns a single product by ID; 404 for inactive or non-existent
- [x] **Stock Adjustment** — Admin: set absolute stock value via `PATCH` endpoint
- [x] **Auto-Increment IDs** — MongoDB does not natively provide auto-increment; implemented via a `counters` collection and `findOneAndUpdate` with `$inc` and upsert — generates sequential `int64` IDs compatible with the protobuf contract
- [x] **MongoDB Indexes** — Ensured at startup via `EnsureIndexes`: `categories.slug` (unique), `products.slug` (unique), `products.sku` (unique), `products.category_id` (non-unique)
- [x] **Dual Interface: gRPC + REST** — REST (port 8081) for admin/public HTTP access; gRPC (port 50051) for internal monolith-to-service calls

**Cart** *(Monolith — Redis)*
- [x] **Add Item** — Validates product existence via gRPC to Product Service; uses `HINCRBY` to increment if already in cart; refreshes 7-day TTL
- [x] **Update Item** — Updates quantity; automatically delegates to remove if quantity ≤ 0
- [x] **Remove Item** — Removes a single product from the Redis hash (`HDEL`)
- [x] **Clear Cart** — Deletes the entire cart key from Redis (`DEL`)
- [x] **Get Cart** — Fetches raw quantities from Redis, then enriches with a single bulk gRPC call (`GetByIDs`); computes `subtotal` and `grand_total`; surfaces `unavailable_items` for products removed from catalog

**Order / Checkout** *(Monolith — PostgreSQL + cross-service gRPC + RabbitMQ)*
- [x] **Batched Stock Decrease** — All stock decrements are sent in a single `BulkDecreaseStock` gRPC call with a list of `{product_id, quantity}` items — no N+1 cross-service calls per cart item
- [x] **Atomic Stock Deduction (MongoDB Session Transaction)** — `BulkDecreaseStock` on Product Service runs inside a MongoDB multi-document session transaction via `session.WithTransaction`; filter `stock >= quantity` prevents oversell without explicit locking — if any item fails, the entire transaction rolls back
- [x] **Saga Compensation (Ledger-Based Idempotent Rollback)** — A `checkoutID` (UUID) is generated before the transaction. If `BulkDecreaseStock` succeeds but `CreateOrderWithItems` fails, `BulkRestoreStock(checkoutID)` is called on Product Service. Restore reads items from the `stock_ledgers` collection and writes a `:restore` ledger entry — subsequent retries are no-ops because the ledger entry already exists
- [x] **Invoice Snapshot** — `OrderItem` stores `product_name` and `price` at time of purchase (`INV-YYYYMMDD-NNNNNN`); immutable to future catalog changes
- [x] **Order History** — `GET /orders`: two queries (orders + items `WHERE order_id IN (?)`); avoids N+1
- [x] **Order Detail** — `GET /orders/:order_id`: ownership check returns `404` rather than `403` to avoid leaking order existence
- [x] **Async Notification Event** — After a successful checkout, `PublishOrderCreated` sends an `order.created` event to RabbitMQ (`order.events` exchange, topic type, routing key `order.created`); publish failure is logged but does **not** roll back or fail the checkout response

**Notification Consumer** *(Standalone service — PostgreSQL pgx, RabbitMQ AMQP)*
- [x] **Event Consumption** — Subscribes to queue `notification.order.created`; QoS prefetch=1 (processes one message at a time)
- [x] **Manual Ack/Nack** — Parse failures: `Nack(requeue=false)` (discard malformed messages); processing failures: `Nack(requeue=true)` (retry); success: `Ack`
- [x] **Idempotent Activity Log** — Writes to `activity_logs` table with `ON CONFLICT (order_id) DO NOTHING` — redelivered messages cannot duplicate log entries
- [x] **Simulated Notification** — Logs a human-readable "email sent" line per event; actual email/push delivery is a future integration point

**Infrastructure & Quality**
- [x] **Docker Compose** — Single-command local environment: 7 services (monolith app, product-service, notification-consumer, PostgreSQL 16, MongoDB 8 replica set, Redis 7, RabbitMQ 4) in an isolated bridge network
- [x] **MongoDB Replica Set** — Required for multi-document session transactions; `mongo-init-replica` one-shot container runs `rs.initiate()` idempotently at startup
- [x] **Multi-stage Dockerfiles** — Each service has its own Dockerfile; builder stage (Go Alpine) produces minimal runtime images
- [x] **Versioned SQL Migrations** — `golang-migrate`-compatible `.up.sql`/`.down.sql` pairs for PostgreSQL (users, orders, order_items, activity_logs tables)
- [x] **Unit Tests — User Module** — 4 test functions covering `Login`, `Register`, `Logout`, `GetProfile`
- [x] **Unit Tests — Product Service** — 10 test functions covering `CreateProduct`, `UpdateProduct`, `DeleteProduct`, `SearchProducts`, `GetProductDetail`, `AdjustStock`, `GetByProductID`, `GetProductsByIDs`, `BulkDecreaseStock`, `BulkRestoreStock`; 3 test functions covering `CreateCategory`, `GetAllCategories`, `ValidateCategoryExists`
- [x] **Unit Tests — Order Module** — 3 test functions, 20 sub-cases covering: success checkout, cart error, empty cart, product lookup error, zero-qty items, product not in map, bulk decrease error, order creation failure + restore success, order creation failure + restore failure, transaction error, clear cart error, publish event failure still succeeds; history with items, history empty, history error; detail success, not found, DB error on FindByID, wrong ownership, items DB error
- [x] **Integration Tests — Product Repository** — 3 test functions (build tag `integration`, uses `testcontainers-go` to spin up MongoDB replica set): partial failure rolls back all stock, idempotent restore (called twice only restores once), concurrent overlapping checkouts produce correct final stock without deadlock
- [x] **Mockery Configuration** — `.mockery.yml` registers all interfaces for consistent mock generation
- [x] **Dependency Injection via Wire** — Compile-time DI wiring for the monolith; injection errors surface at `go generate` not at runtime
- [x] **Swagger / OpenAPI Docs** — Handler functions in Product Service have `@Router` annotations; served at `/swagger/*any` on port 8081
- [x] **Structured Logging** — Logrus per-request middleware: `method`, `path`, `status`, `latency_ms`, `client_ip`
- [x] **Standardized JSON Response** — `ResponseSuccess` / `ResponseError` envelope helpers
- [x] **Graceful Shutdown** — Handles `SIGINT`/`SIGTERM` with 5-second drain window in all services

---

## 🏗️ Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| **Go** | 1.26.4 | Primary language — all services |
| **Gin** | v1.12.0 | HTTP router (monolith + product service) |
| **GORM** | v1.31.2 | ORM for monolith (PostgreSQL) |
| **PostgreSQL** | 16 (Alpine) | Relational store — users, orders, activity_logs |
| **pgx** | v5.10.0 | Native PostgreSQL driver — used directly in notification-consumer (no ORM) |
| **MongoDB** | 8 | Document store — products, categories, stock_ledgers |
| **mongo-driver** | v2.8.0 | Official MongoDB Go driver (v2 API) |
| **Redis** | 7 (Alpine) | JWT session store + cart hash storage |
| **go-redis** | v9 | Redis client |
| **gRPC** | v1.83.0 | Internal RPC between monolith and product-service |
| **protobuf** | v1.36.12 | Contract definition for gRPC (`proto/product/v1/product.proto`) |
| **RabbitMQ** | 4 (management) | Async event bus — `order.created` event |
| **amqp091-go** | v1.13.0 | RabbitMQ AMQP client |
| **JWT (HS256)** | golang-jwt v5.3.1 | Stateless tokens with `user_id` and `role` claims |
| **Wire** | v0.7.0 | Compile-time dependency injection (monolith) |
| **Viper** | v1.21.0 | Config management from `.env` + environment variables |
| **Logrus** | v1.9.4 | Structured, leveled logging |
| **golang-migrate** | — | Versioned PostgreSQL schema migrations |
| **Mockery** | — | Auto-generated interface mocks |
| **testcontainers-go** | v0.44.0 | Docker-based integration tests (MongoDB) |
| **Docker Compose** | — | Local development orchestration |
| **bcrypt** | x/crypto | Secure password hashing |

### Clean Architecture Layers

```
HTTP/gRPC Request → Handler → Usecase → Repository → Database / Redis / gRPC Client
      ↑                                                           |
      └──────────────── DTO (Response) ──────────────────────────┘
```

| Layer | Responsibility |
|---|---|
| **Handler** | HTTP/gRPC concerns: request binding, input validation, status code mapping |
| **Usecase** | Business logic; communicates with other modules/services via interfaces only |
| **Repository** | Data access abstraction — DB, Redis, or gRPC client behind the same interface |
| **Entity** | Plain domain structs; no HTTP tags, no business logic, no cross-module fields |
| **Model** | Database-specific structs (e.g., BSON tags for MongoDB) — never exposed outside repository |

---

## 🔄 Architecture Overview

### JWT Session Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant U as UserUsecase
    participant DB as PostgreSQL
    participant J as pkg/jwt
    participant R as Redis

    Note over C,R: LOGIN
    C->>H: POST /api/v1/login
    H->>U: Login(ctx, request)
    U->>DB: FindByEmail(email)
    DB-->>U: User{ id, email, password_hash, role }
    U->>U: bcrypt.CompareHashAndPassword()
    U->>J: CreateToken({ user_id, role })
    J-->>U: signed JWT (HS256, 30d)
    U->>R: SET auth:session:TOKEN EX 2592000
    H-->>C: 200 · { token }

    Note over C,R: AUTHENTICATED REQUEST
    C->>H: GET /api/v1/profile · Bearer TOKEN
    H->>J: ParseToken — validate signature + expiry
    H->>R: EXISTS auth:session:TOKEN
    R-->>H: 1 (exists)
    H-->>C: 200 · profile data

    Note over C,R: LOGOUT
    C->>H: DELETE /api/v1/logout · Bearer TOKEN
    H->>U: Logout(ctx, token)
    U->>R: DEL auth:session:TOKEN
    H-->>C: 200 · logged out
    Note right of R: Token is immediately invalid for all future requests.
```

### Cart Flow (Redis Hash)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as CartHandler
    participant U as CartUsecase
    participant PS as ProductService (gRPC)
    participant R as Redis

    Note over C,R: ADD ITEM
    C->>H: POST /api/v1/cart · { product_id, quantity }
    H->>U: AddToCart(userID, productID, quantity)
    U->>PS: GetByID(productID) — gRPC call to product-service:50051
    PS-->>U: Product{ id, price, stock }
    U->>R: HINCRBY cart:{userID} {productID} {qty}
    U->>R: EXPIRE cart:{userID} 7d
    H-->>C: 200 · item added

    Note over C,R: GET CART — one Redis call + one batched gRPC call
    C->>H: GET /api/v1/cart
    H->>U: GetCartDetail(userID)
    U->>R: HGETALL cart:{userID}
    R-->>U: map[ product_id → quantity ]
    U->>PS: GetByIDs([ ids... ]) — one batched gRPC call
    PS-->>U: []Product
    U->>U: compute subtotal + grand_total · detect unavailable_items
    H-->>C: 200 · CartDetailResponse
```

### Checkout Flow (Distributed)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as OrderHandler
    participant U as OrderUsecase
    participant CS as CartService (Redis)
    participant PS as ProductService (gRPC)
    participant Mongo as MongoDB (Product Service)
    participant PG as PostgreSQL (Monolith)
    participant MQ as RabbitMQ

    C->>H: POST /api/v1/orders · Bearer TOKEN
    H->>U: Checkout(ctx, userID)
    U->>CS: GetRawCart(userID) — HGETALL from Redis
    CS-->>U: map[ product_id → quantity ]
    U->>PS: GetByIDs([ ids... ]) — one batched gRPC call
    PS-->>U: []Product{ id, name, price, stock }

    Note over U: Generate checkoutID = uuid.NewString()

    Note over U,PG: BEGIN PostgreSQL Transaction
    U->>PS: BulkDecreaseStock(checkoutID, [{product_id, qty}...]) — one gRPC call
    Note over PS,Mongo: MongoDB session.WithTransaction
    PS->>Mongo: InsertOne stock_ledgers (checkoutID:decrease) — idempotent guard
    loop for each item
        PS->>Mongo: UpdateOne products WHERE _id=id AND stock>=qty · $inc stock -qty
        Note right of Mongo: No explicit lock needed — atomic conditional update
    end
    PS-->>U: OK (or ErrInsufficientStock / ErrProductNotFound)

    U->>PG: CreateOrderWithItems(order, items)
    PG-->>U: order.ID · invoice_number = INV-YYYYMMDD-NNNNNN
    Note over U,PG: COMMIT PostgreSQL Transaction

    U->>CS: ClearCart(userID) — DEL cart:{userID}
    U->>MQ: PublishOrderCreated(event) — fire-and-forget
    Note right of MQ: Publish failure is logged but does NOT fail the response.
    H-->>C: 201 · OrderResponse{ order_id, invoice_number, total_amount, items }
```

### Saga Compensation Flow (Stock Restore on Order Failure)

```mermaid
sequenceDiagram
    autonumber
    participant U as OrderUsecase
    participant PS as ProductService (gRPC)
    participant Mongo as MongoDB (Product Service)
    participant PG as PostgreSQL (Monolith)

    Note over U: checkoutID already generated
    U->>PS: BulkDecreaseStock(checkoutID, items) — SUCCESS
    Note over PS,Mongo: stock_ledgers: checkoutID:decrease inserted

    U->>PG: CreateOrderWithItems(...) — FAILS (DB error)
    Note over U,PG: PostgreSQL transaction rolls back

    Note over U: stock already deducted in MongoDB — compensation needed
    U->>PS: BulkRestoreStock(checkoutID) — separate context (5s timeout)
    Note over PS,Mongo: Check if checkoutID:restore already exists in ledger
    PS->>Mongo: FindOne stock_ledgers WHERE _id = checkoutID:restore
    Mongo-->>PS: ErrNoDocuments (not yet restored)
    PS->>Mongo: FindOne stock_ledgers WHERE _id = checkoutID:decrease
    Mongo-->>PS: ledger with original items
    Note over PS,Mongo: MongoDB session.WithTransaction
    loop for each item in decrease ledger
        PS->>Mongo: UpdateOne products · $inc stock +qty
    end
    PS->>Mongo: InsertOne stock_ledgers (checkoutID:restore)
    Note right of Mongo: Future calls with same checkoutID are no-ops (restore ledger exists)
    PS-->>U: OK
    U-->>U: return original error to caller
```

### Notification Event Flow (Async)

```mermaid
sequenceDiagram
    autonumber
    participant U as OrderUsecase
    participant MQ as RabbitMQ
    participant NC as NotificationConsumer
    participant PG as PostgreSQL (activity_logs)

    Note over U,MQ: After successful checkout + cart clear
    U->>MQ: Publish to exchange=order.events · key=order.created · JSON payload
    Note right of U: Publish failure is logged, checkout still returns 201

    Note over MQ,NC: Queue: notification.order.created (durable, topic binding)
    MQ->>NC: Deliver message (QoS prefetch=1)
    NC->>NC: json.Unmarshal(msg.Body) → OrderCreatedEvent
    alt Parse failure
        NC->>MQ: Nack(requeue=false) — discard malformed message
    else Processing
        NC->>NC: Log simulated email notification
        NC->>PG: INSERT INTO activity_logs ... ON CONFLICT (order_id) DO NOTHING
        alt DB insert success
            NC->>MQ: Ack
        else DB insert failure
            NC->>MQ: Nack(requeue=true) — retry
        end
    end
    Note right of PG: ON CONFLICT ensures idempotency — redelivery cannot duplicate the log
```

### Token Revocation Approaches

| Approach | Lookup Cost | Instant Revocation | Notes |
|---|---|---|---|
| Pure Stateless JWT | O(0) | ❌ Waits for expiry | No server-side state |
| DB Blacklist | O(log n) | ✅ | Can become a DB bottleneck at scale |
| **Redis Session (this project)** | **O(1)** | **✅ On DEL** | Redis cluster-ready; cart reuses same client |

---

## 📁 Directory Structure

```
e-commerce/                          # Monorepo root
│
├── docker-compose.yml               # 7 services: app, product-service, notification-consumer,
│                                    #   postgres, mongo (+ init replica), redis, rabbitmq
├── .env / .env.example              # Shared environment config
├── .mockery.yml                     # Mockery config: all interfaces across all modules
├── go.mod / go.sum                  # Single workspace go.mod (monorepo with replace directives)
│
├── proto/
│   └── product/v1/
│       └── product.proto            # gRPC contract: GetByID, GetByIDs, BulkDecreaseStock,
│                                    #   BulkRestoreStock — shared by monolith and product-service
│
├── pkg/                             # Shared libraries (used by all services in the monorepo)
│   ├── apperror/                    # Typed sentinel errors (ErrProductNotFound, ErrInsufficientStock, …)
│   ├── cache/                       # Redis client factory
│   ├── config/                      # Config struct + Viper loader (.env → struct)
│   ├── database/                    # PostgreSQL (pgx pool) + MongoDB client factories
│   ├── engine/                      # Gin engine setup
│   ├── jwt/                         # JwtToken interface + HS256 impl; Auth{ UserID, Role }
│   ├── logger/                      # Logrus wrapper
│   ├── messaging/
│   │   ├── topology.go              # RabbitMQ constants + DeclareOrderEventsTopology()
│   │   │                            #   Exchange: order.events (topic, durable)
│   │   │                            #   Queue:    notification.order.created (durable)
│   │   │                            #   Binding:  routing key = order.created
│   │   └── rabbitmq.go              # AMQP connection factory
│   ├── middleware/                  # AuthMiddleware, AdminMiddleware (shared across services)
│   ├── response/                    # ResponseSuccess / ResponseError envelope helpers
│   ├── skugen/                      # SKU auto-generation
│   ├── transaction/                 # WithTransaction(ctx, fn) — isolates TX boilerplate
│   └── validator/                   # go-playground/validator singleton
│
├── monolith/                        # Monolith service (User · Cart · Order)
│   ├── Dockerfile
│   ├── cmd/api/
│   │   ├── main.go                  # Entry point: --seed flag, graceful shutdown
│   │   ├── wire.go                  # Wire provider declarations (build tag wireinject)
│   │   └── wire_gen.go              # Auto-generated DI (do not edit)
│   ├── database/
│   │   ├── migration/               # Timestamp-prefixed .up.sql / .down.sql (PostgreSQL)
│   │   │   ├── *_create_users_table.up.sql
│   │   │   ├── *_create_orders_table.up.sql
│   │   │   ├── *_create_order_items_table.up.sql
│   │   │   └── *_create_activity_logs_table.up.sql
│   │   └── seeder/                  # Seeds admin user + sample data (--seed flag)
│   ├── dependency/                  # Infrastructure adapters (monolith-specific)
│   │   ├── app.go                   # App struct wiring
│   │   ├── gorm.go                  # GORM on top of pgx pool (Postgres)
│   │   ├── grpc.go                  # gRPC client conn to product-service:50051
│   │   └── rabbitmq.go              # AMQP channel factory + topology declaration
│   └── internal/
│       ├── user/                    # MODULE: Auth & User Identity
│       │   ├── entity/user.go       # User{ ID, Name, Email, Password, Role }
│       │   ├── dto/                 # Request/Response structs with validator tags
│       │   ├── repository/          # FindByEmail, FindByID, Create (GORM + Redis)
│       │   ├── mocks/
│       │   └── usecase/
│       │       ├── user_usecase.go
│       │       ├── user_usecase_impl.go   # bcrypt, JWT, Redis session lifecycle
│       │       └── user_usecase_test.go   # 4 test functions
│       │
│       ├── product/                 # MODULE: ProductService interface (gRPC client stub)
│       │   ├── entity/product.go    # Product{ … } + BulkDecreaseStock{ ProductID, Quantity }
│       │   ├── mocks/               # MockProductService
│       │   ├── repository/
│       │   │   └── product_grpc_client.go  # ProductGRPCClient: implements ProductService
│       │   │                               #   interface by translating to proto calls
│       │   └── usecase/
│       │       └── contract.go      # ProductService interface: GetByProductID, GetProductsByIDs,
│       │                            #   BulkDecreaseStock, BulkRestoreStock
│       │
│       ├── cart/                    # MODULE: Shopping Cart (Redis-backed)
│       │   ├── dto/                 # CartItemCreateRequest; CartDetailResponse{ Items, UnavailableItems, GrandTotal }
│       │   ├── repository/          # HINCRBY / HSET / HDEL / HGETALL / DEL
│       │   └── usecase/
│       │       ├── cart_usecase.go
│       │       ├── cart_usecase_impl.go    # Bulk enrichment via gRPC, unavailable_items detection
│       │       └── contract.go             # CartService{ GetRawCart, ClearCart }
│       │
│       ├── order/                   # MODULE: Checkout & Order History
│       │   ├── entity/
│       │   │   ├── order.go         # Order{ ID, UserID, InvoiceNumber, TotalAmount, Status }
│       │   │   └── order_item.go    # OrderItem{ … ProductName, Price — snapshot at checkout }
│       │   ├── dto/
│       │   ├── event/
│       │   │   └── order_created_event.go  # OrderCreatedEvent{ OrderID, UserID, InvoiceNumber, TotalAmount }
│       │   ├── repository/
│       │   │   ├── order_repository.go
│       │   │   ├── order_repository_impl.go   # CreateOrderWithItems: insert → INV-YYYYMMDD-NNNNNN → insert items
│       │   │   └── notification_event_publisher.go  # NotificationEventPublisher: publishes to RabbitMQ
│       │   ├── mocks/
│       │   ├── delivery/http/
│       │   └── usecase/
│       │       ├── order_usecase.go
│       │       ├── order_usecase_impl.go   # Checkout saga: GetRawCart → GetByIDs → BulkDecreaseStock
│       │       │                           #   → CreateOrderWithItems (with compensation) → ClearCart → Publish
│       │       ├── event_publisher.go      # EventPublisher interface
│       │       └── order_usecase_test.go   # 3 test functions, 20 sub-cases
│       │
│       ├── mocks/                   # Shared mocks: MockRedis, MockTransaction
│       └── routes/router.go         # Route groups: public · protected (no admin — product routes moved to product-service)
│
└── service/
    ├── product-service/             # Microservice: Product Catalog
    │   ├── Dockerfile
    │   ├── cmd/main.go              # Entry: MongoDB + Redis + gRPC server (50051) + HTTP server (8081)
    │   └── internal/product/
    │       ├── entity/              # Product{ … } + Category{ … } (pure domain)
    │       ├── model/               # BSON-tagged structs: ProductModel, CategoryModel, StockLedgerModel
    │       ├── repository/
    │       │   ├── mongo_shared.go            # getNextSequence (counters collection) + EnsureIndexes
    │       │   ├── product_repository_impl.go # CRUD + BulkDecreaseStock + BulkRestoreStock
    │       │   │                              #   (both use session.WithTransaction for atomicity)
    │       │   ├── category_repository_impl.go
    │       │   ├── session_repository.go      # Redis session (auth middleware reuse)
    │       │   └── product_repository_integration_test.go  # build tag: integration
    │       ├── usecase/             # 10 + 3 test functions
    │       ├── dto/
    │       ├── mocks/
    │       └── delivery/
    │           ├── http/            # Gin handlers for REST API (port 8081)
    │           └── grpc/
    │               └── product_grpc_server.go  # gRPC server: GetByID, GetByIDs,
    │                                           #   BulkDecreaseStock, BulkRestoreStock
    │
    └── notification-consumer/       # Microservice: Async Notification Consumer
        ├── Dockerfile
        ├── cmd/                     # Entry: PostgreSQL (pgx pool) + RabbitMQ channel + consumer loop
        ├── dependency/              # Infrastructure adapters for this service
        └── internal/notification/
            ├── entity/              # ActivityLog{ ID, OrderID, UserID, Message, CreatedAt }
            ├── dto/                 # OrderCreatedEvent (matches publisher payload)
            ├── repository/
            │   └── activity_log_repository_impl.go  # pgx native (no ORM)
            │                                         #   INSERT ... ON CONFLICT (order_id) DO NOTHING
            ├── usecase/
            │   └── notification_usecase_impl.go  # HandleOrderCreated: log + write activity_log
            └── delivery/consumer/
                └── order_created_consumer_impl.go  # Start(): Qos(1) → Consume → manual Ack/Nack loop
```

### Service Boundary Rules

**Within the monolith** — modules communicate via `contract.go` interfaces only:
```
✅ ALLOWED:   internal/cart  → internal/product/usecase  (via ProductService interface)
✅ ALLOWED:   internal/order → internal/cart/usecase     (via CartService interface)
✅ ALLOWED:   internal/order → internal/product/usecase  (via ProductService interface)
❌ FORBIDDEN: any module     → another module's repository (direct data access)
❌ FORBIDDEN: any module     → another module's entity     (direct struct import)
```

**Across service boundaries** — the protobuf contract is the only allowed communication surface:
```
✅ ALLOWED:   monolith/product/repository → Product Service    (via gRPC, productv1.ProductServiceClient)
✅ ALLOWED:   monolith/order/repository   → RabbitMQ           (via messaging.NotificationEventPublisher)
✅ ALLOWED:   notification-consumer       → RabbitMQ           (via AMQP consumer)
❌ FORBIDDEN: monolith → Product Service's MongoDB directly
❌ FORBIDDEN: monolith → Product Service's internal packages
```

The `ProductService` interface in `monolith/internal/product/usecase/contract.go` is the same interface that was previously satisfied by a local GORM/MySQL implementation. The migration to gRPC required only swapping the implementation registered in `wire.go` — usecase and handler code were untouched.

---

## 📡 API Endpoints

### Authentication & User *(Monolith — port 8080)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/register` | Register a new user | — |
| `POST` | `/api/v1/login` | Login and receive JWT session token | — |
| `GET` | `/api/v1/profile` | Get authenticated user profile | ✅ JWT |
| `DELETE` | `/api/v1/logout` | Revoke current JWT session | ✅ JWT |

### Cart *(Monolith — port 8080)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api/v1/cart` | View cart with enriched product data | ✅ JWT |
| `POST` | `/api/v1/cart` | Add item to cart | ✅ JWT |
| `PATCH` | `/api/v1/cart/:product_id` | Update item quantity | ✅ JWT |
| `DELETE` | `/api/v1/cart/:product_id` | Remove single item | ✅ JWT |
| `DELETE` | `/api/v1/cart` | Clear entire cart | ✅ JWT |

### Orders *(Monolith — port 8080)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/orders` | Checkout cart into a new order | ✅ JWT |
| `GET` | `/api/v1/orders` | List authenticated user's order history | ✅ JWT |
| `GET` | `/api/v1/orders/:order_id` | Get single order detail | ✅ JWT |

### Product Catalog *(Product Service — port 8081)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api/v1/categories` | List all categories | — |
| `POST` | `/api/v1/admin/categories` | Create category (auto slug) | ✅ Admin |
| `GET` | `/api/v1/products` | List & search products with pagination | — |
| `GET` | `/api/v1/products/:product_id` | Get product detail | — |
| `POST` | `/api/v1/admin/products` | Create product | ✅ Admin |
| `PATCH` | `/api/v1/admin/products/:product_id` | Update product | ✅ Admin |
| `DELETE` | `/api/v1/admin/products/:product_id` | Soft-delete product | ✅ Admin |
| `PATCH` | `/api/v1/admin/products/:product_id/adjust-stock` | Set absolute stock value | ✅ Admin |

### Internal gRPC *(Product Service — port 50051)*

| RPC | Description |
|---|---|
| `GetByID(GetByIDRequest)` | Fetch single product by ID |
| `GetByIDs(GetByIDsRequest)` | Batch fetch products by IDs |
| `BulkDecreaseStock(BulkDecreaseStockRequest)` | Atomically decrease stock for multiple products (idempotent via ledger) |
| `BulkRestoreStock(BulkRestoreStockRequest)` | Compensate a previously failed decrease (idempotent via ledger) |

---

## 📖 API Documentation

Product Service handler functions have Swagger annotations generated via `swaggo/swag`.

**Access the UI:**
1. Start all services: `docker compose up -d`
2. Open [http://localhost:8081/swagger/index.html](http://localhost:8081/swagger/index.html) — Product Catalog API
3. Monolith (User/Cart/Order) API is accessible at `http://localhost:8080`

**Testing authenticated endpoints:** Click the **Authorize** button in Swagger UI and enter the Bearer token returned by `POST /api/v1/login` on the monolith (port 8080). The same JWT is valid on both services (shared secret key).

---

## 📦 Request & Response Samples

### `POST /api/v1/register` (port 8080)

```json
// Request
{ "name": "Budi Santoso", "email": "budi@example.com", "password": "securepassword123" }

// 201 Created
{ "success": true, "message": "user registered successfully", "data": { "id": 1, "name": "Budi Santoso", "email": "budi@example.com" } }

// 409 Conflict
{ "success": false, "message": "Email Already Exists" }
```

### `POST /api/v1/login` (port 8080)

```json
// Request
{ "email": "budi@example.com", "password": "securepassword123" }

// 200 OK
{ "success": true, "message": "user logged in successfully", "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." } }

// 401 Unauthorized
{ "success": false, "message": "Wrong Email or Password" }
```

### `GET /api/v1/products?search=keyboard&category_id=1&page=1&limit=10` (port 8081)

```json
// 200 OK
{
  "success": true,
  "message": "Product fetched successfully",
  "data": {
    "data": [
      {
        "id": 12, "category_id": 1, "name": "Mechanical Keyboard 60%",
        "slug": "mechanical-keyboard-60", "price": 750000, "stock": 50,
        "sku": "KBD-60-001", "is_active": true
      }
    ],
    "meta": { "page": 1, "limit": 10, "total": 1 }
  }
}
```

### `POST /api/v1/cart` & `GET /api/v1/cart` (port 8080)

```json
// POST Request — Authorization: Bearer <token>
{ "product_id": 12, "quantity": 2 }
// 200 OK
{ "success": true, "message": "item added to cart successfully" }

// GET Response
{
  "success": true, "message": "cart detail retrieved successfully",
  "data": {
    "items": [{ "product_id": 12, "name": "Mechanical Keyboard 60%", "price": 750000, "quantity": 2, "subtotal": 1500000, "stock_available": 50 }],
    "unavailable_items": [],
    "grand_total": 1500000
  }
}
```

> **`unavailable_items`**: Products that exist in Redis but have since been deactivated or deleted from the catalog. Surfaced for user review rather than silently removed.

### `POST /api/v1/orders` (Checkout — port 8080)

```json
// Request — Authorization: Bearer <token>
// No request body needed; cart is resolved server-side from JWT user_id

// 201 Created
{
  "success": true,
  "message": "Checkout successful",
  "data": {
    "order_id": 6,
    "invoice_number": "INV-20260710-000006",
    "total_amount": 200000,
    "status": "PAID",
    "items": [
      { "product_id": 1, "product_name": "Produk A", "price": 10000, "quantity": 10, "subtotal": 100000 },
      { "product_id": 2, "product_name": "Produk B", "price": 20000, "quantity": 5,  "subtotal": 100000 }
    ]
  }
}

// 400 Bad Request — cart is empty
{ "success": false, "message": "Cart Is Empty" }

// 404 Not Found — a product in cart was removed from catalog
{ "success": false, "message": "Product Not Found" }

// 422 Unprocessable Entity — stock insufficient for one or more items
{ "success": false, "message": "Insufficient Stock" }
```

---

## 🚀 Getting Started

### Option A: Docker Compose (Recommended)

This starts **7 services** in an isolated bridge network: monolith app, product-service, notification-consumer, PostgreSQL 16, MongoDB 8, Redis 7, RabbitMQ 4. No local database setup required.

**1. Clone the repository**
```bash
git clone https://github.com/Mpayy/e-commerce.git
cd e-commerce
```

**2. Configure environment variables**
```bash
cp .env.example .env
# Edit .env with your preferred values (JWT secret, etc.)
```

**3. Start all services**
```bash
docker compose up --build -d
```

> MongoDB replica set initialization is handled automatically by the `mongo-init-replica` one-shot container — no manual `rs.initiate()` required.

**4. Run database migrations** (PostgreSQL — users, orders, order_items, activity_logs)
```bash
docker exec -it ecommerce-app sh -c \
  "migrate -database 'postgres://postgres:postgres@postgres:5432/ecommerce?sslmode=disable' -path database/migration up"
```

**5. (Optional) Seed initial data**
```bash
docker exec -it ecommerce-app ./main --seed
```

**6. Verify services are running**
```bash
# Monolith (User/Cart/Order)
curl -s http://localhost:8080/api/v1/products | jq .  # will proxy via gRPC to product-service

# Product Service (direct)
curl -s http://localhost:8081/health

# RabbitMQ Management UI
open http://localhost:15672  # guest / guest
```

**Stop all services:**
```bash
docker compose down
```

---

### Option B: Manual Setup

**Prerequisites**

| Tool | Minimum Version |
|---|---|
| Go | 1.22+ |
| PostgreSQL | 16+ |
| MongoDB | 8+ (with replica set for transactions) |
| Redis | 7.0+ |
| RabbitMQ | 4+ |
| `golang-migrate` CLI | latest |

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**1. Clone & install dependencies**
```bash
git clone https://github.com/Mpayy/e-commerce.git
cd e-commerce
go mod download
```

**2. Configure environment**
```bash
cp .env.example .env
```

`.env` reference:
```env
# PostgreSQL (monolith + notification-consumer)
DATABASE_HOST=127.0.0.1
DATABASE_PORT=5432
DATABASE_NAME=ecommerce
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=postgres

# MongoDB (product-service)
MONGODB_URI=mongodb://127.0.0.1:27017/ecommerce?replicaSet=rs0
MONGODB_DATABASE=ecommerce

# Redis (monolith + product-service)
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_DB=0

# RabbitMQ (monolith publisher + notification-consumer)
RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/

# gRPC — monolith to product-service
PRODUCT_SERVICE_ADDR=127.0.0.1:50051

APP_HOST=0.0.0.0
APP_PORT=8080
LOG_LEVEL=debug

# Use a long, random string in production
JWT_SECRET_KEY=change-me-to-a-strong-random-secret-key
```

**3. Initialize MongoDB replica set** (required for transactions)
```bash
mongosh --eval 'rs.initiate({ _id: "rs0", members: [{ _id: 0, host: "127.0.0.1:27017" }] })'
```

**4. Run PostgreSQL migrations**
```bash
migrate -database "postgres://postgres:postgres@127.0.0.1:5432/ecommerce?sslmode=disable" \
        -path monolith/database/migration up
```

**5. Start services** (3 separate terminals)
```bash
# Terminal 1 — Product Service (gRPC :50051 + HTTP :8081)
go run ./service/product-service/cmd/...

# Terminal 2 — Notification Consumer
go run ./service/notification-consumer/cmd/...

# Terminal 3 — Monolith (HTTP :8080)
go run ./monolith/cmd/api/...
```

---

### Developer Commands

**Regenerate Wire DI** (after modifying `wire.go` in the monolith):
```bash
cd monolith/cmd/api && go generate ./...
```

**Regenerate mocks** (after modifying any interface):
```bash
go generate ./...
```

**Regenerate Swagger docs** (after modifying handler annotations in product-service):
```bash
swag init -g service/product-service/cmd/main.go --output service/product-service/internal/product/docs
```

**Run unit tests** (no live dependencies required — all mocked):
```bash
go test ./monolith/... ./service/product-service/internal/product/usecase/... -v -race
```

**Run integration tests** (requires Docker — spins up a real MongoDB replica set via testcontainers):
```bash
go test -tags=integration ./service/product-service/internal/product/repository/... -v
```

**Run tests for a specific module:**
```bash
# User module
go test -v ./monolith/internal/user/usecase/... -run "TestUserUsecaseImpl"

# Product service usecases
go test -v ./service/product-service/internal/product/usecase/... -run "TestProductUsecaseImpl"

# Category usecases
go test -v ./service/product-service/internal/product/usecase/... -run "TestCategoryUsecaseImpl"

# Order module
go test -v ./monolith/internal/order/usecase/... -run "TestOrderUsecase"
```

---

## 📝 Architecture Notes

### Why Product Catalog Was Extracted First

Product is the leaf node in the dependency graph: User, Cart, and Order all depend on it, but Product depends on nothing else. This made it the highest-leverage extraction candidate — removing it from the monolith severs the most cross-cutting dependencies in one step. Extracting User or Order first would have left Product as shared state accessed by both the remaining monolith and the new service, creating a worse coupling problem than the one being solved.

### Why MongoDB for Products, PostgreSQL for Everything Else

Products have inherently flexible, schema-variable attributes (different categories need different fields). Forcing that into a relational schema produces sparse columns or EAV anti-patterns. MongoDB's document model fits naturally. In contrast, users, orders, and activity logs have strict, predictable schemas where ACID guarantees and strong consistency are requirements — PostgreSQL is the correct tool there. Running two databases is a tradeoff accepted deliberately, not an oversight.

### Preventing Oversell: Row Lock vs. Atomic Conditional Update

In the old relational world, preventing oversell required `SELECT ... FOR UPDATE` to hold a row-level lock until the transaction commits. MongoDB's approach is different: `updateOne` with a filter of `{ stock: { $gte: quantity } }` and `$inc: { stock: -quantity }` is atomic at the document level. If the filter doesn't match (stock too low), `MatchedCount == 0` and the operation is rejected — no explicit lock, no waiting goroutines, no deadlock possible from the locking side.

### Why MongoDB Multi-Document Transactions Don't Deadlock the Same Way

Relational deadlocks arise from circular lock waits: transaction A holds row X and waits for row Y, while transaction B holds row Y and waits for row X. MongoDB's transaction model uses optimistic concurrency with `WriteConflict` errors: if two transactions touch the same document simultaneously, one receives a `WriteConflict` and is automatically retried by `session.WithTransaction`. There is no "waiting" — the conflict is detected immediately and the retry happens from scratch. This eliminates the circular-wait condition that causes relational deadlocks, at the cost of making conflicting transactions slower under high contention (retries vs. waiting).

### Saga Pattern for Stock Consistency

The checkout flow uses a minimal saga: `BulkDecreaseStock` (MongoDB, cross-service) happens before `CreateOrderWithItems` (PostgreSQL, local). If the PostgreSQL step fails, the MongoDB decrement has already committed. The compensation (`BulkRestoreStock`) is called immediately with the same `checkoutID`. Idempotency is enforced by the `stock_ledgers` collection: the decrease operation inserts a `checkoutID:decrease` document (duplicate key → no-op), and the restore operation inserts `checkoutID:restore` (already exists → no-op). Network retries of either operation are safe because the ledger prevents double-application.

### Why `order.created` Is RabbitMQ, Not Another gRPC Call

gRPC is synchronous: the caller waits for a response before returning to the user. A notification side-effect that fails should not cause the checkout to fail, and it should not add latency to the critical path. RabbitMQ decouples the two: the monolith publishes and moves on. If the notification consumer is down, messages queue up and are processed when it recovers. A gRPC call to a notification service would couple availability: if the notification service is unreachable, checkout fails — an unacceptable dependency for a non-critical side effect.

### Cart Enrichment Without N+1

The `GET /cart` flow keeps database/service round trips O(1) regardless of cart size:

1. **One Redis call**: `HGETALL cart:{userID}` returns the full `map[product_id → quantity]`
2. **One gRPC call**: `GetByIDs([ids...])` fetches all products in a single request to Product Service
3. Everything else (subtotal, grand_total, unavailable_items detection) is computed in memory

### Transaction Abstraction

`pkg/transaction.WithTransaction(ctx, fn)` wraps PostgreSQL transaction boilerplate, keeping it out of usecase logic. The transaction is stored in the context and extracted in repository methods via `GetTxFromContext`. This also allows the transaction itself to be mocked in unit tests — usecases remain testable without a live database connection.

---

<div align="center">
<sub>MIT License · Contributions welcome</sub>
</div>
