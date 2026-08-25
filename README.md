<div align="center">

# 🛒 E-Commerce API

### Go · Gin · sqlx · sqlc · PostgreSQL · MongoDB · Redis · gRPC · RabbitMQ · JWT · Docker

[![CI Pipeline](https://img.shields.io/github/actions/workflow/status/Mpayy/e-commerce/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&label=CI%20Pipeline)](https://github.com/Mpayy/e-commerce/actions/workflows/ci.yml)
[![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=for-the-badge)](https://github.com/Mpayy/e-commerce)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

<br />

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12.0-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://github.com/gin-gonic/gin)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![MongoDB](https://img.shields.io/badge/MongoDB-8-47A248?style=for-the-badge&logo=mongodb&logoColor=white)](https://www.mongodb.com/)
[![Redis](https://img.shields.io/badge/Redis-v9-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![gRPC](https://img.shields.io/badge/gRPC-v1.83.0-244C5A?style=for-the-badge&logo=grpc&logoColor=white)](https://grpc.io/)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-4-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)](https://www.rabbitmq.com/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docs.docker.com/compose/)
[![Swagger](https://img.shields.io/badge/Swagger-OpenAPI-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)](http://localhost:8081/swagger/index.html)

<p align="center">
  <em>A full microservices e-commerce backend in Go: five independently deployable services (API Gateway, User, Product, Order &amp; Cart, Notification Consumer) communicating via HTTP reverse-proxy, gRPC, and RabbitMQ.<br>
  Each service owns its own database and explores a different data-access technology — sqlx, sqlc, mongo-driver, pgx native — structured on Clean Architecture with explicit cross-service boundaries enforced by protobuf contracts and a saga-based compensation pattern for distributed stock consistency.</em>
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

**Authentication & User** *(User Service — PostgreSQL via sqlx, port 8082)*
- [x] **User Registration** — Password hashing with bcrypt; domain-specific error on duplicate email
- [x] **JWT Authentication (HS256)** — 30-day tokens; secret key loaded from environment config via Viper
- [x] **Redis Session Management** — JWT stored in Redis (`auth:session:<token>`) on login; atomically deleted on logout for immediate invalidation
- [x] **Auth Middleware** — Validates Bearer token signature, expiry, and Redis session existence; injects `auth` context
- [x] **Role-Based Access Control (RBAC)** — Admin middleware reads role from JWT claims; returns `403` for unauthorized access
- [x] **User Profile** — Returns authenticated user data resolved from `user_id` in the JWT
- [x] **sqlx Queries** — Hand-written SQL with `sqlx.DB.GetContext` / `QueryRowxContext`; `BeginTxx` for transactions when needed; no ORM

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
- [x] **Dual Interface: gRPC + REST** — REST (port 8081) for admin/public HTTP access; gRPC (port 50051) for internal Order-Service-to-Product-Service calls

**Cart & Shopping** *(Order & Cart Service — Redis, port 8083)*
- [x] **Add Item** — Validates product existence via gRPC to Product Service; uses `HINCRBY` to increment if already in cart; refreshes 7-day TTL
- [x] **Update Item** — Updates quantity; automatically delegates to remove if quantity ≤ 0
- [x] **Remove Item** — Removes a single product from the Redis hash (`HDEL`)
- [x] **Clear Cart** — Deletes the entire cart key from Redis (`DEL`)
- [x] **Get Cart** — Fetches raw quantities from Redis, then enriches with a single bulk gRPC call (`GetByIDs`); computes `subtotal` and `grand_total`; surfaces `unavailable_items` for products removed from catalog

**Order / Checkout** *(Order & Cart Service — PostgreSQL via sqlc, port 8083)*
- [x] **Batched Stock Decrease** — All stock decrements are sent in a single `BulkDecreaseStock` gRPC call with a list of `{product_id, quantity}` items — no N+1 cross-service calls per cart item
- [x] **Atomic Stock Deduction (MongoDB Session Transaction)** — `BulkDecreaseStock` on Product Service runs inside a MongoDB multi-document session transaction via `session.WithTransaction`; filter `stock >= quantity` prevents oversell without explicit locking — if any item fails, the entire transaction rolls back
- [x] **Saga Compensation (Ledger-Based Idempotent Rollback)** — A `checkoutID` (UUID) is generated before the transaction. If `BulkDecreaseStock` succeeds but `CreateOrderWithItems` fails, `BulkRestoreStock(checkoutID)` is called on Product Service. Restore reads items from the `stock_ledgers` collection and writes a `:restore` ledger entry — subsequent retries are no-ops because the ledger entry already exists
- [x] **Invoice Snapshot** — `OrderItem` stores `product_name` and `price` at time of purchase (`INV-YYYYMMDD-NNNNNN`); immutable to future catalog changes
- [x] **Order History with Pagination** — `GET /orders?page=N&limit=N`: two queries (orders with `LIMIT`/`OFFSET` + items `WHERE order_id IN (?)`); avoids N+1; returns total count for pagination metadata
- [x] **Order Detail** — `GET /orders/:order_id`: ownership check returns `404` rather than `403` to avoid leaking order existence
- [x] **PostgreSQL Transaction via sqlc** — `CreateOrderWithItems` opens a `pgxpool` transaction, calls `queries.WithTx(tx)` from generated sqlc code, inserts order + items + invoice number atomically, then commits; transaction management lives in the repository, not the usecase
- [x] **Async Notification Event** — After a successful checkout, `PublishOrderCreated` sends an `order.created` event to RabbitMQ (`order.events` exchange, topic type, routing key `order.created`); publish failure is logged but does **not** roll back or fail the checkout response

**Notification Consumer** *(Standalone service — PostgreSQL pgx native, RabbitMQ AMQP)*
- [x] **Event Consumption** — Subscribes to queue `notification.order.created`; QoS prefetch=1 (processes one message at a time)
- [x] **Manual Ack/Nack** — Parse failures: `Nack(requeue=false)` (discard malformed messages); processing failures: `Nack(requeue=true)` (retry); success: `Ack`
- [x] **Idempotent Activity Log** — Writes to `activity_logs` table with `ON CONFLICT (order_id) DO NOTHING` — redelivered messages cannot duplicate log entries
- [x] **Simulated Notification** — Logs a human-readable "email sent" line per event; actual email/push delivery is a future integration point

**API Gateway** *(port 8080 — sole public entry point)*
- [x] **Reverse Proxy** — Built with Go's `net/http/httputil.ReverseProxy`; no external tools (Kong/Traefik); path-prefix-based routing sorted by prefix length (longest-match first) to avoid ambiguity
- [x] **Path Routing Table** — `/api/v1/cart` → Order Service; `/api/v1/orders` → Order Service; `/api/v1/products` + `/api/v1/categories` + `/api/v1/admin/*` → Product Service; `/api/v1/register`, `/api/v1/login`, `/api/v1/profile`, `/api/v1/logout` → User Service
- [x] **Passthrough Proxy** — Gateway does NOT perform authentication; `Authorization` header and all other headers are forwarded verbatim to the target service, which validates JWT + Redis session independently

**Infrastructure & Quality**
- [x] **Docker Compose** — Single-command local environment: 9 containers (api-gateway, user-service, product-service, order-service, notification-consumer, PostgreSQL 16, MongoDB 8 replica set, Redis 7, RabbitMQ 4) in an isolated bridge network
- [x] **PostgreSQL Multi-Database Init** — `scripts/init-multiple-postgres-databases.sh` auto-creates 3 logical databases (`user_db`, `order_db`, `notification_db`) on first container start; each service owns its own database
- [x] **MongoDB Replica Set** — Required for multi-document session transactions; `mongo-init-replica` one-shot container runs `rs.initiate()` idempotently at startup
- [x] **Multi-stage Dockerfiles** — Each service has its own Dockerfile; builder stage (Go Alpine) produces minimal runtime images
- [x] **Versioned SQL Migrations** — Separate `golang-migrate`-compatible `.up.sql`/`.down.sql` per service: `user-service/database/migration/` (users table), `order-service/database/migration/` (orders, order_items tables), `notification-consumer/database/migration/` (activity_logs table)
- [x] **Unit Tests — User Service** — 4 test functions covering `Register`, `Login`, `Logout`, `GetProfile`
- [x] **Unit Tests — Product Service** — 10 test functions covering `CreateProduct`, `UpdateProduct`, `DeleteProduct`, `SearchProducts`, `GetProductDetail`, `AdjustStock`, `GetByProductID`, `GetProductsByIDs`, `BulkDecreaseStock`, `BulkRestoreStock`; 3 test functions covering `CreateCategory`, `GetAllCategories`, `ValidateCategoryExists`
- [x] **Unit Tests — Cart Module** — 6 test functions covering `AddToCart`, `UpdateCartItem`, `RemoveFromCart`, `GetCartDetail`, `GetRawCart`, `ClearCart`
- [x] **Unit Tests — Order Module** — 3 test functions, 20 sub-cases covering: success checkout, cart error, empty cart, product lookup error, zero-qty items, product not in map, bulk decrease error, order creation failure + restore success, order creation failure + restore failure, clear cart error, publish event failure still succeeds; history with items, history empty, history error, pagination; detail success, not found, DB error on FindByID, wrong ownership, items DB error
- [x] **Integration Tests — Product Repository** — 3 test functions (build tag `integration`, uses `testcontainers-go` to spin up MongoDB replica set): partial failure rolls back all stock, idempotent restore (called twice only restores once), concurrent overlapping checkouts produce correct final stock without deadlock
- [x] **Mockery Configuration** — `.mockery.yml` registers all interfaces for consistent mock generation
- [x] **Swagger / OpenAPI Docs** — Handler functions in User Service, Product Service, and Order & Cart Service have `@Router` annotations; served at `/swagger/*any` on their respective ports (`:8082`, `:8081`, `:8083`)
- [x] **Structured Logging** — Logrus per-request middleware: `method`, `path`, `status`, `latency_ms`, `client_ip`
- [x] **Standardized JSON Response** — `ResponseSuccess` / `ResponseError` envelope helpers
- [x] **Graceful Shutdown** — Handles `SIGINT`/`SIGTERM` with 5-second drain window in all services

---

## 🏗️ Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| **Go** | 1.26.4 | Primary language — all services |
| **Gin** | v1.12.0 | HTTP router (user-service, product-service, order-service) |
| **sqlx** | v1.3.5 | SQL query helper in user-service — typed scanning, `GetContext`, `QueryRowxContext`; no ORM |
| **sqlc** | v2 | Type-safe Go code generated from SQL queries — used in order-service (`pgx/v5` backend) |
| **pgx** | v5.10.0 | Native PostgreSQL driver — connection pool in order-service (sqlc backend) + direct use in notification-consumer (no ORM or helper layer) |
| **PostgreSQL** | 16 (Alpine) | Relational store — 3 logical DBs: `user_db`, `order_db`, `notification_db` |
| **MongoDB** | 8 | Document store — products, categories, stock_ledgers |
| **mongo-driver** | v2.8.0 | Official MongoDB Go driver (v2 API) — used directly in product-service |
| **Redis** | 7 (Alpine) | JWT session store (user-service + middleware) + cart hash storage (order-service) |
| **go-redis** | v9 | Redis client |
| **gRPC** | v1.83.0 | Internal RPC — order-service → product-service (stock operations, product lookups) |
| **protobuf** | v1.36.12 | Contract definition for gRPC (`proto/product/v1/product.proto`) |
| **RabbitMQ** | 4 (management) | Async event bus — `order.created` event after checkout |
| **amqp091-go** | v1.13.0 | RabbitMQ AMQP client |
| **JWT (HS256)** | golang-jwt v5.3.1 | Stateless tokens with `user_id` and `role` claims |
| **Viper** | v1.21.0 | Config management from `.env` + environment variables |
| **Logrus** | v1.9.4 | Structured, leveled logging |
| **golang-migrate** | — | Versioned PostgreSQL schema migrations (separate paths per service) |
| **Mockery** | — | Auto-generated interface mocks |
| **testcontainers-go** | v0.44.0 | Docker-based integration tests (MongoDB replica set) |
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

### Request Flow via API Gateway

```mermaid
flowchart LR
    C([Client])

    subgraph GW["API Gateway :8080 (net/http reverse proxy)"]
        direction TB
        R["Path-Prefix Router (longest-match first)"]
    end

    subgraph US["User Service :8082"]
        UA["Auth / Profile — sqlx → PostgreSQL user_db"]
    end

    subgraph PS["Product Service :8081"]
        PA["Product / Category — mongo-driver → MongoDB"]
    end

    subgraph OS["Order & Cart Service :8083"]
        OA["Cart + Order — sqlc/Redis → PostgreSQL order_db"]
    end

    C -->|"All public traffic :8080"| GW
    GW -->|"/api/v1/register /api/v1/login /api/v1/profile /api/v1/logout"| US
    GW -->|"/api/v1/products /api/v1/categories /api/v1/admin/..."| PS
    GW -->|"/api/v1/cart /api/v1/orders"| OS
```

> The gateway performs **pure passthrough**: no authentication, no request transformation. The `Authorization` header is forwarded verbatim. Each downstream service validates JWT + Redis session independently (defense in depth).

### JWT Session Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant GW as API Gateway :8080
    participant H as UserService Handler :8082
    participant U as UserUsecase
    participant DB as PostgreSQL (user_db)
    participant J as pkg/jwt
    participant R as Redis

    Note over C,R: LOGIN
    C->>GW: POST /api/v1/login
    GW->>H: proxy passthrough (no auth check)
    H->>U: Login(ctx, request)
    U->>DB: FindByEmail(email) — sqlx.GetContext
    DB-->>U: User{ id, email, password_hash, role }
    U->>U: bcrypt.CompareHashAndPassword()
    U->>J: CreateToken({ user_id, role })
    J-->>U: signed JWT (HS256, 30d)
    U->>R: SET auth:session:TOKEN EX 2592000
    H-->>C: 200 · { token }

    Note over C,R: AUTHENTICATED REQUEST
    C->>GW: GET /api/v1/profile · Bearer TOKEN
    GW->>H: proxy passthrough (Authorization header forwarded as-is)
    H->>J: ParseToken — validate signature + expiry
    H->>R: EXISTS auth:session:TOKEN
    R-->>H: 1 (exists)
    H-->>C: 200 · profile data

    Note over C,R: LOGOUT
    C->>GW: DELETE /api/v1/logout · Bearer TOKEN
    GW->>H: proxy passthrough
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
    participant GW as API Gateway :8080
    participant H as CartHandler (Order Service :8083)
    participant U as CartUsecase
    participant PS as ProductService (gRPC :50051)
    participant R as Redis

    Note over C,R: ADD ITEM
    C->>GW: POST /api/v1/cart · { product_id, quantity }
    GW->>H: proxy to order-service:8083
    H->>U: AddToCart(userID, productID, quantity)
    U->>PS: GetByID(productID) — gRPC call to product-service:50051
    PS-->>U: Product{ id, price, stock }
    U->>R: HINCRBY cart:{userID} {productID} {qty}
    U->>R: EXPIRE cart:{userID} 7d
    H-->>C: 200 · item added

    Note over C,R: GET CART — one Redis call + one batched gRPC call
    C->>GW: GET /api/v1/cart
    GW->>H: proxy to order-service:8083
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
    participant GW as API Gateway :8080
    participant H as OrderHandler (Order Service :8083)
    participant U as OrderUsecase
    participant CS as CartService (Redis)
    participant PS as ProductService (gRPC :50051)
    participant Mongo as MongoDB (Product Service)
    participant PG as "PostgreSQL order_db (sqlc + pgxpool)"
    participant MQ as RabbitMQ

    C->>GW: POST /api/v1/orders · Bearer TOKEN
    GW->>H: proxy to order-service:8083
    H->>U: Checkout(ctx, userID)
    U->>CS: GetRawCart(userID) — HGETALL from Redis
    CS-->>U: map[ product_id → quantity ]
    U->>PS: GetByIDs([ ids... ]) — one batched gRPC call
    PS-->>U: []Product{ id, name, price, stock }

    Note over U: Generate checkoutID = uuid.NewString()

    U->>PS: BulkDecreaseStock(checkoutID, items) — one gRPC call
    Note over PS,Mongo: MongoDB session.WithTransaction
    PS->>Mongo: InsertOne stock_ledgers (checkoutID:decrease) — idempotent guard
    loop for each item
        PS->>Mongo: UpdateOne products WHERE stock>=qty · $inc stock -qty
    end
    PS-->>U: OK (or ErrInsufficientStock / ErrProductNotFound)

    Note over U,PG: repository.CreateOrderWithItems — opens pgxpool tx internally
    U->>PG: qtx.CreateOrder → qtx.UpdateInvoiceNumber → qtx.CreateOrderItem x N
    PG-->>U: order.ID · invoice_number = INV-YYYYMMDD-NNNNNN
    Note over PG: tx.Commit() — sqlc WithTx pattern, managed in repository layer

    U->>CS: ClearCart(userID) — DEL cart:{userID}
    U->>MQ: PublishOrderCreated(event) — fire-and-forget
    Note right of MQ: Publish failure is logged but does NOT fail the response.
    H-->>C: 201 · OrderResponse{ order_id, invoice_number, total_amount, items }
```

### Saga Compensation Flow (Stock Restore on Order Failure)

```mermaid
sequenceDiagram
    autonumber
    participant U as OrderUsecase (Order Service)
    participant PS as ProductService (gRPC)
    participant Mongo as MongoDB (Product Service)
    participant PG as PostgreSQL order_db

    Note over U: checkoutID already generated
    U->>PS: BulkDecreaseStock(checkoutID, items) — SUCCESS
    Note over PS,Mongo: stock_ledgers: checkoutID:decrease inserted

    U->>PG: CreateOrderWithItems(...) — FAILS (DB error)
    Note over PG: pgxpool transaction rolls back automatically (deferred Rollback)

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
    participant U as OrderUsecase (Order Service)
    participant MQ as RabbitMQ
    participant NC as NotificationConsumer
    participant PG as "PostgreSQL (notification_db · pgx native)"

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
e-commerce/                          # Monorepo root (single go.mod)
│
├── docker-compose.yml               # 9 containers: api-gateway, user-service, product-service,
│                                    #   order-service, notification-consumer, postgres, mongo
│                                    #   (+ init replica), redis, rabbitmq
├── scripts/
│   └── init-multiple-postgres-databases.sh  # Creates user_db, order_db, notification_db on
│                                            #   first postgres container start
├── .env / .env.example              # Shared environment config (DATABASE_NAME overridden per service)
├── .mockery.yml                     # Mockery config: all interfaces across all services
├── go.mod / go.sum                  # Single workspace go.mod (monorepo)
│
├── proto/
│   └── product/v1/
│       └── product.proto            # gRPC contract: GetByID, GetByIDs, BulkDecreaseStock,
│                                    #   BulkRestoreStock — shared by order-service and product-service
│
├── pkg/                             # Shared libraries (used by all services)
│   ├── apperror/                    # Typed sentinel errors (ErrProductNotFound, ErrInsufficientStock, …)
│   ├── cache/                       # Redis client factory
│   ├── config/                      # Config struct + Viper loader (.env → struct); single
│   │                                #   DatabaseName field overridden per service via compose env
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
│   └── validator/                   # go-playground/validator singleton
│
└── services/
    ├── api-gateway/                 # Entry point: path-based HTTP reverse proxy (no CA layers)
    │   ├── Dockerfile
    │   ├── cmd/main.go              # Entry: loads routes from config, starts net/http on :8080
    │   └── internal/gateway/
    │       ├── config/route.go      # BuildRoutes(): path-prefix → service address mapping
    │       │                        #   Sorted by prefix length (longest first) to avoid ambiguity
    │       └── proxy/
    │           └── reverse_proxy.go # Gateway struct: httputil.NewSingleHostReverseProxy per target;
    │                                #   ServeHTTP iterates routes, proxies on first prefix match
    │
    ├── user-service/                # Microservice: Auth & User Identity (PostgreSQL via sqlx)
    │   ├── Dockerfile
    │   ├── cmd/main.go              # Entry: sqlx pool + Redis + Gin; HTTP :8082; Swagger docs
    │   ├── database/
    │   │   ├── migration/           # 20260628_create_users_table.up/down.sql
    │   │   └── seeder/              # Seeds admin user (--seed flag)
    │   ├── dependency/
    │   │   ├── postgres.go          # pgxpool factory
    │   │   └── sqlx.go              # sqlx.DB wrapping pgxpool for sqlx-style scanning
    │   ├── docs/                    # Swagger generated output (swagger.json, docs.go)
    │   └── internal/user/
    │       ├── entity/user.go       # User{ ID, Name, Email, Password, Role }
    │       ├── dto/                 # Request/Response structs with validator tags
    │       ├── repository/          # FindByEmail, FindByID, Create — hand-written SQL via sqlx
    │       ├── mocks/
    │       └── usecase/
    │           ├── user_usecase.go
    │           ├── user_usecase_impl.go   # bcrypt, JWT, Redis session lifecycle
    │           └── user_usecase_test.go   # 4 test functions
    │
    ├── order-service/               # Microservice: Shopping Cart + Checkout & Orders
    │   │                            # (PostgreSQL via sqlc + Redis cart; port 8083)
    │   ├── Dockerfile
    │   ├── sqlc.yaml                # sqlc config: queries → internal/order/repository/sqlc/
    │   ├── cmd/main.go              # Entry: pgxpool + Redis + gRPC conn + RabbitMQ + Gin; :8083
    │   ├── database/
    │   │   ├── migration/           # 20260823_create_orders_table.up/down.sql
    │   │   │                        # 20260823_create_order_items_table.up/down.sql
    │   │   └── query/               # .sql files consumed by sqlc (ListOrdersByUser,
    │   │                            #   CountOrdersByUser, ListItemsByOrderIDs, GetOrderByID,
    │   │                            #   GetItemsByOrderID, CreateOrder, UpdateInvoiceNumber, CreateOrderItem)
    │   ├── docs/                    # Swagger generated output
    │   └── internal/
    │       ├── cart/                # MODULE: Shopping Cart (Redis-backed)
    │       │   ├── dto/             # CartItemCreateRequest; CartDetailResponse
    │       │   ├── repository/      # HINCRBY / HSET / HDEL / HGETALL / DEL
    │       │   ├── mocks/
    │       │   └── usecase/
    │       │       ├── cart_usecase_impl.go    # Bulk enrichment via gRPC, unavailable_items
    │       │       ├── cart_usecase_test.go    # 6 test functions
    │       │       └── contract.go             # CartService{ GetRawCart, ClearCart, … }
    │       │
    │       ├── order/               # MODULE: Checkout & Order History (sqlc-generated queries)
    │       │   ├── entity/
    │       │   │   ├── order.go     # Order{ ID, UserID, InvoiceNumber, TotalAmount, Status }
    │       │   │   └── order_item.go  # OrderItem{ … ProductName, Price — snapshot at checkout }
    │       │   ├── dto/
    │       │   ├── event/
    │       │   │   └── order_created_event.go  # OrderCreatedEvent{ … }
    │       │   ├── repository/
    │       │   │   ├── order_repository_impl.go  # CreateOrderWithItems: opens pgxpool tx,
    │       │   │   │                             #   queries.WithTx(tx), inserts + commits
    │       │   │   │                             # FindByUserID: LIMIT/OFFSET + CountOrdersByUser
    │       │   │   ├── event_publisher_impl.go   # NotificationEventPublisher → RabbitMQ
    │       │   │   └── sqlc/                     # Auto-generated by sqlc: db.go, models.go,
    │       │   │                                 #   order.sql.go, order_item.sql.go
    │       │   ├── mocks/
    │       │   ├── delivery/http/
    │       │   └── usecase/
    │       │       ├── order_usecase_impl.go   # Checkout saga: no WithTransaction at usecase level;
    │       │       │                           #   transaction entirely managed in repository
    │       │       └── order_usecase_test.go   # 3 test functions, 20 sub-cases
    │       │
    │       └── product/             # MODULE: ProductService interface (gRPC client stub)
    │           ├── entity/product.go
    │           ├── mocks/
    │           └── repository/
    │               └── product_grpc_client.go  # Translates ProductService interface to proto calls
    │
    ├── product-service/             # Microservice: Product Catalog
    │   ├── Dockerfile
    │   ├── cmd/main.go              # Entry: MongoDB + Redis + gRPC server (:50051) + HTTP (:8081)
    │   └── internal/product/
    │       ├── entity/              # Product + Category (pure domain)
    │       ├── model/               # BSON-tagged structs: ProductModel, CategoryModel, StockLedgerModel
    │       ├── repository/
    │       │   ├── mongo_shared.go            # getNextSequence + EnsureIndexes
    │       │   ├── product_repository_impl.go # CRUD + BulkDecreaseStock + BulkRestoreStock
    │       │   │                              #   (both use session.WithTransaction)
    │       │   ├── category_repository_impl.go
    │       │   ├── session_repository.go      # Redis session (auth middleware reuse)
    │       │   └── product_repository_integration_test.go  # build tag: integration
    │       ├── usecase/             # 10 + 3 test functions
    │       ├── dto/
    │       ├── mocks/
    │       └── delivery/
    │           ├── http/            # Gin handlers for REST API (port 8081)
    │           └── grpc/
    │               └── product_grpc_server.go  # gRPC: GetByID, GetByIDs, BulkDecreaseStock, BulkRestoreStock
    │
    └── notification-consumer/       # Microservice: Async Notification Consumer
        ├── Dockerfile
        ├── cmd/                     # Entry: PostgreSQL (pgx pool) + RabbitMQ + consumer loop
        ├── database/
        │   └── migration/           # 20260817_create_activity_logs_table.up/down.sql
        ├── dependency/
        └── internal/notification/
            ├── entity/              # ActivityLog{ ID, OrderID, UserID, Message, CreatedAt }
            ├── dto/                 # OrderCreatedEvent (matches publisher payload)
            ├── repository/
            │   └── activity_log_repository_impl.go  # pgx native (no ORM, no helper library)
            │                                         #   INSERT ... ON CONFLICT (order_id) DO NOTHING
            ├── usecase/
            │   └── notification_usecase_impl.go  # HandleOrderCreated: log + write activity_log
            └── delivery/consumer/
                └── order_created_consumer_impl.go  # Start(): Qos(1) → Consume → manual Ack/Nack
```

### Service Boundary Rules

**Within order-service** — `cart` and `order` modules coexist in one process and communicate via interfaces:
```
✅ ALLOWED:   internal/order → internal/cart/usecase    (via CartService interface)
✅ ALLOWED:   internal/order → internal/product/usecase (via ProductService interface)
✅ ALLOWED:   internal/cart  → internal/product/usecase (via ProductService interface)
❌ FORBIDDEN: any module     → another module's repository (direct data access)
❌ FORBIDDEN: any module     → another module's entity     (direct struct import)
```
> `CartService` is consumed exclusively by `OrderUsecase`. Keeping them in the same service avoids a gratuitous network hop for a tightly coupled flow (checkout reads cart → checks out → clears cart).

**Across service boundaries** — the protobuf contract and HTTP proxy are the only allowed communication surfaces:
```
✅ ALLOWED:   api-gateway            → any service          (HTTP reverse proxy, path-prefix routing)
✅ ALLOWED:   order-service/product  → Product Service      (via gRPC, productv1.ProductServiceClient)
✅ ALLOWED:   order-service/order    → RabbitMQ             (via messaging.NotificationEventPublisher)
✅ ALLOWED:   notification-consumer  → RabbitMQ             (via AMQP consumer)
❌ FORBIDDEN: order-service → Product Service's MongoDB directly
❌ FORBIDDEN: order-service → Product Service's internal packages
❌ FORBIDDEN: api-gateway   → any service's internal packages or DB
```

---

## 📡 API Endpoints

> All endpoints below are accessible via the **API Gateway at port 8080** (e.g., `http://localhost:8080/api/v1/login`). The gateway performs path-based routing to the appropriate service. Direct service ports are also available for development/testing.

### Authentication & User *(User Service — direct: port 8082 · via gateway: port 8080)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/register` | Register a new user | — |
| `POST` | `/api/v1/login` | Login and receive JWT session token | — |
| `GET` | `/api/v1/profile` | Get authenticated user profile | ✅ JWT |
| `DELETE` | `/api/v1/logout` | Revoke current JWT session | ✅ JWT |

### Cart *(Order & Cart Service — direct: port 8083 · via gateway: port 8080)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api/v1/cart` | View cart with enriched product data | ✅ JWT |
| `POST` | `/api/v1/cart` | Add item to cart | ✅ JWT |
| `PATCH` | `/api/v1/cart/:product_id` | Update item quantity | ✅ JWT |
| `DELETE` | `/api/v1/cart/:product_id` | Remove single item | ✅ JWT |
| `DELETE` | `/api/v1/cart` | Clear entire cart | ✅ JWT |

### Orders *(Order & Cart Service — direct: port 8083 · via gateway: port 8080)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/orders` | Checkout cart into a new order | ✅ JWT |
| `GET` | `/api/v1/orders?page=1&limit=10` | List authenticated user's order history (paginated) | ✅ JWT |
| `GET` | `/api/v1/orders/:order_id` | Get single order detail | ✅ JWT |

### Product Catalog *(Product Service — direct: port 8081 · via gateway: port 8080)*

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

### Internal gRPC *(Product Service — port 50051, not exposed outside Docker network)*

| RPC | Description |
|---|---|
| `GetByID(GetByIDRequest)` | Fetch single product by ID |
| `GetByIDs(GetByIDsRequest)` | Batch fetch products by IDs |
| `BulkDecreaseStock(BulkDecreaseStockRequest)` | Atomically decrease stock for multiple products (idempotent via ledger) |
| `BulkRestoreStock(BulkRestoreStockRequest)` | Compensate a previously failed decrease (idempotent via ledger) |

---

## 📖 API Documentation

Each HTTP-facing service documents its own API independently via `swaggo/swag`. The API Gateway and the Notification Consumer are excluded — the gateway is a generic reverse proxy with no per-endpoint handlers to annotate; the consumer has no HTTP API at all.

**Access the Swagger UI:**

| Service | Swagger UI | Notes |
|---|---|---|
| User Service | http://localhost:8082/swagger/index.html | Auth, registration, profile, logout |
| Product Catalog Service | http://localhost:8081/swagger/index.html | Products, categories, stock management |
| Order & Cart Service | http://localhost:8083/swagger/index.html | Cart operations + checkout + order history |

**Testing authenticated endpoints:** Get a Bearer token from `POST /api/v1/login` (on port 8080, 8082, or 8083 — they all route through the same User Service). Click **Authorize** in any Swagger UI and paste it in. The token works across all services because they share `JWT_SECRET_KEY` and the same Redis session store.

---

## 📦 Request & Response Samples

### `POST /api/v1/register` (port 8080 or 8082)

```json
// Request
{ "name": "Budi Santoso", "email": "budi@example.com", "password": "securepassword123" }

// 201 Created
{ "success": true, "message": "user registered successfully", "data": { "id": 1, "name": "Budi Santoso", "email": "budi@example.com" } }

// 409 Conflict
{ "success": false, "message": "Email Already Exists" }
```

### `POST /api/v1/login` (port 8080 or 8082)

```json
// Request
{ "email": "budi@example.com", "password": "securepassword123" }

// 200 OK
{ "success": true, "message": "user logged in successfully", "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." } }

// 401 Unauthorized
{ "success": false, "message": "Wrong Email or Password" }
```

### `GET /api/v1/products?search=keyboard&category_id=1&page=1&limit=10` (port 8080 or 8081)

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

### `POST /api/v1/cart` & `GET /api/v1/cart` (port 8080 or 8083)

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

### `POST /api/v1/orders` (Checkout — port 8080 or 8083)

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

This starts **9 containers** in an isolated bridge network: api-gateway, user-service, product-service, order-service, notification-consumer, PostgreSQL 16, MongoDB 8, Redis 7, RabbitMQ 4. No local database setup required.

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

> **PostgreSQL databases** (`user_db`, `order_db`, `notification_db`) are created automatically by `scripts/init-multiple-postgres-databases.sh` on first container start — no manual `CREATE DATABASE` required.

> **MongoDB replica set** initialization is handled automatically by the `mongo-init-replica` one-shot container — no manual `rs.initiate()` required.

**4. Run database migrations**

Migrations are per-service with separate paths and databases:

```bash
# user_db — users table
docker exec -it ecommerce-user-service sh -c \
  "migrate -database 'postgres://postgres:postgres@postgres:5432/user_db?sslmode=disable' \
           -path database/migration up"

# order_db — orders + order_items tables
docker exec -it ecommerce-order-service sh -c \
  "migrate -database 'postgres://postgres:postgres@postgres:5432/order_db?sslmode=disable' \
           -path database/migration up"

# notification_db — activity_logs table
docker exec -it ecommerce-notification-consumer sh -c \
  "migrate -database 'postgres://postgres:postgres@postgres:5432/notification_db?sslmode=disable' \
           -path database/migration up"
```

**5. (Optional) Seed initial admin user**
```bash
docker exec -it ecommerce-user-service ./main --seed
```

**6. Verify services are running**
```bash
# API Gateway (single entry point)
curl -s http://localhost:8080/health

# User Service (direct)
curl -s http://localhost:8082/health

# Product Service (direct)
curl -s http://localhost:8081/health

# Order Service (direct)
curl -s http://localhost:8083/health

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
# PostgreSQL (shared host/port; database name overridden per service via environment variable)
DATABASE_HOST=127.0.0.1
DATABASE_PORT=5432
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=postgres

# MongoDB (product-service)
MONGODB_URI=mongodb://127.0.0.1:27017/ecommerce?replicaSet=rs0
MONGODB_DATABASE=ecommerce

# Redis (user-service session + order-service cart)
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_DB=0

# RabbitMQ (order-service publisher + notification-consumer)
RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/

# gRPC — order-service to product-service
PRODUCT_SERVICE_ADDR=127.0.0.1:50051

# Gateway service addresses (used by api-gateway only)
USER_SERVICE_ADDR=http://127.0.0.1:8082
PRODUCT_SERVICE_HTTP_ADDR=http://127.0.0.1:8081
ORDER_SERVICE_ADDR=http://127.0.0.1:8083

APP_HOST=0.0.0.0
LOG_LEVEL=debug

# Use a long, random string in production
JWT_SECRET_KEY=change-me-to-a-strong-random-secret-key
```

**3. Create PostgreSQL databases**
```bash
psql -U postgres -c "CREATE DATABASE user_db;"
psql -U postgres -c "CREATE DATABASE order_db;"
psql -U postgres -c "CREATE DATABASE notification_db;"
```

**4. Initialize MongoDB replica set** (required for multi-document transactions)
```bash
mongosh --eval 'rs.initiate({ _id: "rs0", members: [{ _id: 0, host: "127.0.0.1:27017" }] })'
```

**5. Run PostgreSQL migrations** (3 separate commands)
```bash
# user_db
migrate -database "postgres://postgres:postgres@127.0.0.1:5432/user_db?sslmode=disable" \
        -path services/user-service/database/migration up

# order_db
migrate -database "postgres://postgres:postgres@127.0.0.1:5432/order_db?sslmode=disable" \
        -path services/order-service/database/migration up

# notification_db
migrate -database "postgres://postgres:postgres@127.0.0.1:5432/notification_db?sslmode=disable" \
        -path services/notification-consumer/database/migration up
```

**6. Start services** (5 separate terminals)
```bash
# Terminal 1 — Product Service (gRPC :50051 + HTTP :8081)
DATABASE_NAME=ecommerce go run ./services/product-service/cmd/...

# Terminal 2 — User Service (HTTP :8082)
DATABASE_NAME=user_db go run ./services/user-service/cmd/...

# Terminal 3 — Order & Cart Service (HTTP :8083)
DATABASE_NAME=order_db go run ./services/order-service/cmd/...

# Terminal 4 — Notification Consumer
DATABASE_NAME=notification_db go run ./services/notification-consumer/cmd/...

# Terminal 5 — API Gateway (HTTP :8080)
go run ./services/api-gateway/cmd/...
```

---

### Developer Commands

**Regenerate mocks** (after modifying any interface):
```bash
go generate ./...
```

**Regenerate sqlc code** (after modifying SQL queries in `order-service/database/query/`):
```bash
cd services/order-service && sqlc generate
```

**Regenerate Swagger docs** (after modifying handler annotations):
```bash
# User Service
swag init -g services/user-service/cmd/main.go --output services/user-service/docs

# Product Service
swag init -g services/product-service/cmd/main.go --output services/product-service/internal/product/docs

# Order & Cart Service
swag init -g services/order-service/cmd/main.go --output services/order-service/docs
```

**Run all unit tests** (no live dependencies required — all mocked):
```bash
go test ./services/... -v -race
```

**Run tests for a specific service:**
```bash
# User Service
go test -v ./services/user-service/internal/user/usecase/... -run "TestUserUsecaseImpl"

# Product Service usecases
go test -v ./services/product-service/internal/product/usecase/... -run "TestProductUsecaseImpl"

# Category usecases
go test -v ./services/product-service/internal/product/usecase/... -run "TestCategoryUsecaseImpl"

# Cart module
go test -v ./services/order-service/internal/cart/usecase/... -run "TestCartUsecaseImpl"

# Order module
go test -v ./services/order-service/internal/order/usecase/... -run "TestOrderUsecase"
```

**Run integration tests** (requires Docker — spins up a real MongoDB replica set via testcontainers):
```bash
go test -tags=integration ./services/product-service/internal/product/repository/... -v
```

---

## 📝 Architecture Notes

### Why Product Catalog Was Extracted First

Product is the leaf node in the dependency graph: User, Cart, and Order all depend on it, but Product depends on nothing else. This made it the highest-leverage extraction candidate — removing it from the monolith severs the most cross-cutting dependencies in one step. Extracting User or Order first would have left Product as shared state accessed by both the remaining monolith and the new service, creating a worse coupling problem than the one being solved.

### Why User Was Extracted Before Order

User has the lowest extraction risk: no other service calls User synchronously. Its extraction is a pure decomposition exercise — move code, set up a new database, done. Order is the most complex: it acts as an orchestrator (calls Cart via Redis, calls Product via gRPC, publishes to RabbitMQ, writes to its own PostgreSQL). Extracting it last ensured all its dependencies were already stable, independently deployed services before the orchestrator itself moved.

### Why Cart and Order Live in the Same Service

`CartService` is consumed exclusively by `OrderUsecase`. At checkout, `OrderUsecase` calls `cartService.GetRawCart` → computes items → proceeds → calls `cartService.ClearCart`. Splitting Cart and Order into two separate services would add a gratuitous network hop to this single business flow with no architectural benefit: there is no other consumer of Cart, no independent scaling need, and no reason for them to fail independently. The boundary between `cart/` and `order/` modules within the service is still enforced via interfaces — the discipline is there, the process isolation is not necessary.

### Why the API Gateway Is Built from Scratch

Using Go's `net/http/httputil.ReverseProxy` instead of an off-the-shelf tool (Kong, Traefik, NGINX) was a deliberate choice for two reasons. First, consistency: the entire stack is Go; adding an external routing layer would introduce operational complexity without adding capability at this scale. Second, portfolio value: implementing a reverse proxy from scratch demonstrates understanding of how HTTP proxying works (URL rewriting, header forwarding, backend addressing), whereas configuring an existing tool only demonstrates familiarity with its DSL. The implementation is simple by design — it does exactly one thing: path-prefix matching + passthrough.

### Why Authentication Stays in Each Service (Not the Gateway)

Moving JWT validation to the gateway would create a single point of failure for authentication and couple the gateway to the JWT/Redis infrastructure. More importantly, it would make individual services impossible to test or call directly during development without the gateway in the path. Keeping auth at the service level provides defense in depth: even if the gateway is misconfigured or bypassed, each service remains secure. All services share the same `pkg/middleware` (from the monorepo's `pkg/` package), so there is no code duplication — only process distribution.

### Database Technology per Service — Deliberate Heterogeneity

Each service uses a different data-access approach, chosen for its characteristics rather than uniformity:

| Service | Technology | Rationale |
|---|---|---|
| **user-service** | `sqlx` | Simple CRUD on a stable relational schema; sqlx gives typed scanning without the overhead or magic of an ORM |
| **order-service** | `sqlc` | Generates type-safe Go from SQL; ensures query correctness at compile time; `WithTx` pattern integrates cleanly with pgxpool transactions |
| **product-service** | `mongo-driver` (v2) | Products have variable, schema-flexible attributes; MongoDB's document model avoids sparse columns; multi-document transactions via `session.WithTransaction` handle stock operations atomically |
| **notification-consumer** | `pgx` native | Simple fire-and-forget insert; no query complexity warrants a helper library — raw `pool.Exec` is the most direct path |

The original monolith used GORM throughout. After the monolith was fully decomposed, GORM was abandoned, and the data access layer for each service was rewritten using the most appropriate tool for its specific workload.

### Preventing Oversell: Row Lock vs. Atomic Conditional Update

In a relational system, preventing oversell typically requires `SELECT ... FOR UPDATE` to hold a row-level lock until the transaction commits. MongoDB's approach is different: `updateOne` with a filter of `{ stock: { $gte: quantity } }` and `$inc: { stock: -quantity }` is atomic at the document level. If the filter doesn't match (stock too low), `MatchedCount == 0` and the operation is rejected — no explicit lock, no waiting goroutines, no deadlock possible from the locking side.

### Why MongoDB Multi-Document Transactions Don't Deadlock the Same Way

Relational deadlocks arise from circular lock waits: transaction A holds row X and waits for row Y, while transaction B holds row Y and waits for row X. MongoDB's transaction model uses optimistic concurrency with `WriteConflict` errors: if two transactions touch the same document simultaneously, one receives a `WriteConflict` and is automatically retried by `session.WithTransaction`. There is no "waiting" — the conflict is detected immediately and the retry happens from scratch. This eliminates the circular-wait condition that causes relational deadlocks, at the cost of making conflicting transactions slower under high contention (retries vs. waiting).

### Saga Pattern for Stock Consistency

The checkout flow uses a minimal saga: `BulkDecreaseStock` (MongoDB, cross-service) happens before `CreateOrderWithItems` (PostgreSQL, local). If the PostgreSQL step fails, the MongoDB decrement has already committed. The compensation (`BulkRestoreStock`) is called immediately with the same `checkoutID`. Idempotency is enforced by the `stock_ledgers` collection: the decrease operation inserts a `checkoutID:decrease` document (duplicate key → no-op), and the restore operation inserts `checkoutID:restore` (already exists → no-op). Network retries of either operation are safe because the ledger prevents double-application.

### Two PostgreSQL Transaction Patterns: sqlx vs. sqlc

The project uses two different approaches to PostgreSQL transactions across services, each idiomatic to its chosen library:

- **user-service (sqlx)**: Uses `db.BeginTxx(ctx, nil)` to start a transaction, then passes the `*sqlx.Tx` to repository methods explicitly. The transaction handle propagates via method parameters.
- **order-service (sqlc)**: The `Queries` object (generated by sqlc) exposes a `WithTx(tx pgx.Tx) *Queries` method. `CreateOrderWithItems` opens a `pgxpool.Pool.Begin(ctx)` transaction internally in the repository, creates `qtx := queries.WithTx(tx)`, runs all inserts through `qtx`, then commits. Transaction management is fully encapsulated in the repository — the usecase calls a single method and has no visibility into transaction lifecycle.

### Why `order.created` Is RabbitMQ, Not Another gRPC Call

gRPC is synchronous: the caller waits for a response before returning to the user. A notification side-effect that fails should not cause the checkout to fail, and it should not add latency to the critical path. RabbitMQ decouples the two: the order-service publishes and moves on. If the notification consumer is down, messages queue up and are processed when it recovers. A gRPC call to a notification service would couple availability: if the notification service is unreachable, checkout fails — an unacceptable dependency for a non-critical side effect.

### Cart Enrichment Without N+1

The `GET /cart` flow keeps database/service round trips O(1) regardless of cart size:

1. **One Redis call**: `HGETALL cart:{userID}` returns the full `map[product_id → quantity]`
2. **One gRPC call**: `GetByIDs([ids...])` fetches all products in a single request to Product Service
3. Everything else (subtotal, grand_total, unavailable_items detection) is computed in memory

---

<div align="center">
<sub>MIT License · Contributions welcome</sub>
</div>
