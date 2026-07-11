<div align="center">

# 🛒 E-Commerce API

### Go · Gin · GORM · MySQL · Redis · JWT · Docker

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12.0-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://github.com/gin-gonic/gin)
[![GORM](https://img.shields.io/badge/GORM-v1.31.2-E10098?style=for-the-badge)](https://gorm.io/)
[![Redis](https://img.shields.io/badge/Redis-v9-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docs.docker.com/compose/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=for-the-badge)]()

<p align="center">
  <em>A modular monolith e-commerce REST API written in Go, structured with Clean Architecture principles.<br>
  Focused on clear module boundaries, testable business logic, and a straightforward path toward horizontal scalability.</em>
</p>

</div>

---

## 📋 Table of Contents

- [Features & Roadmap](#-features--roadmap)
- [Tech Stack](#-tech-stack)
- [Architecture Overview](#-architecture-overview)
- [Directory Structure](#-directory-structure)
- [API Endpoints](#-api-endpoints)
- [Request & Response Samples](#-request--response-samples)
- [Getting Started](#-getting-started)
- [Architecture Notes](#-architecture-notes)

---

## ✅ Features & Roadmap

### 🟢 Implemented

**Authentication & User**
- [x] **User Registration** — Password hashing with bcrypt; domain-specific error on duplicate email
- [x] **JWT Authentication (HS256)** — 30-day tokens; secret key loaded from environment config via Viper
- [x] **Redis Session Management** — JWT stored in Redis (`auth:session:<token>`) on login; atomically deleted on logout for immediate invalidation
- [x] **Auth Middleware** — Validates Bearer token signature, expiry, and Redis session existence; injects `auth` context
- [x] **Role-Based Access Control (RBAC)** — Admin middleware reads role from JWT claims; returns `403` for unauthorized access
- [x] **User Profile** — Returns authenticated user data resolved from `user_id` in the JWT

**Product Catalog**
- [x] **Category Management** — Admin: create categories with auto-generated slugs, conflict detection on duplicate names
- [x] **Product — Create** — Admin: validates category, auto-generates SKU if not provided, detects duplicate slug/SKU
- [x] **Product — Update** — Admin: full field update including `is_active` toggle; re-validates category only when changed
- [x] **Product — Soft Delete** — Admin: sets `is_active = false` instead of hard deleting
- [x] **Product — Listing & Search** — Public: paginated list with optional `search` (name), `category_id`, `page`, and `limit` query params
- [x] **Product — Detail** — Public: returns a single product by ID; 404 for inactive or non-existent products
- [x] **Stock Adjustment** — Admin: set absolute stock value via `PATCH` endpoint, wrapped in a DB transaction

**Cart**
- [x] **Add Item** — Validates product existence through the `ProductService` interface; uses `HINCRBY` to increment if already in cart; refreshes 7-day TTL
- [x] **Update Item** — Updates quantity; automatically delegates to remove if quantity ≤ 0
- [x] **Remove Item** — Removes a single product from the Redis hash (`HDEL`)
- [x] **Clear Cart** — Deletes the entire cart key from Redis (`DEL`)
- [x] **Get Cart** — Fetches raw quantities from Redis, then enriches with a single bulk DB call (`WHERE id IN (...)`); computes `subtotal` and `grand_total`; surfaces `unavailable_items` for products removed from catalog

**Infrastructure & Quality**
- [x] **Docker Compose** — Single-command local environment: App + MySQL 8.0 + Redis 7 in an isolated bridge network
- [x] **Multi-stage Dockerfile** — Builder stage (Go 1.26 Alpine) produces a minimal Alpine runtime image (~10MB final)
- [x] **Unit Tests — User Module** — 4 test functions covering `Login`, `Register`, `Logout`, `GetProfile` with mockery-generated mocks
- [x] **Unit Tests — Product Module** — 9 test functions covering `CreateProduct`, `UpdateProduct`, `DeleteProduct`, `SearchProducts`, `GetProductDetail`, `AdjustStock`, `GetByProductID`, `GetProductsByIDs`, `DecreaseStock`
- [x] **Unit Tests — Category Module** — 3 test functions covering `CreateCategory`, `GetAllCategories`, `ValidateCategoryExists`
- [x] **Mockery Configuration** — `.mockery.yml` registers all interfaces (repositories, usecases, Redis, JWT, Transaction) for consistent mock generation
- [x] **Cross-Module Contracts** — `ProductService` and `CartService` interfaces defined in `contract.go` files enforce module boundary isolation
- [x] **Structured Logging** — Logrus middleware logs `method`, `path`, `status`, `latency_ms`, `client_ip` per request
- [x] **Standardized JSON Response** — `ResponseSuccess` / `ResponseError` helpers ensure a consistent response envelope
- [x] **Versioned SQL Migrations** — `golang-migrate`-compatible, timestamp-prefixed `.up.sql`/`.down.sql` pairs
- [x] **Graceful Shutdown** — Handles `SIGINT`/`SIGTERM` with a 5-second drain window
- [x] **Dependency Injection via Wire** — Compile-time DI wiring; injection errors surface at `go generate` not at runtime
- [x] **Database Seeder** — `--seed` CLI flag bootstraps admin user and sample categories

### 🔜 Planned

- [ ] **Checkout** — Convert Redis cart to MySQL `orders` + `order_items` in a single ACID transaction; pessimistic stock locking via `SELECT ... FOR UPDATE`
- [ ] **Order History** — `GET /orders` and `GET /orders/:id` for authenticated users
- [ ] **Invoice Snapshot** — `OrderItem` stores `product_name` and `price` at checkout time; order history remains accurate regardless of future catalog changes
- [ ] **Swagger / OpenAPI Docs** — Interactive API documentation via `swaggo`

---

## 🏗️ Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| **Go** | 1.26.4 | Primary language — compiled, goroutine-based concurrency |
| **Gin** | v1.12.0 | HTTP router; radix-tree path matching, request binding, middleware chain |
| **GORM** | v1.31.2 | ORM with parameterized queries and connection pool management |
| **MySQL** | 8.0 | Primary relational store; supports row-level locking (`SELECT FOR UPDATE`) |
| **Redis** | go-redis v9 | JWT session store + cart storage (Redis Hash with TTL) |
| **JWT (HS256)** | golang-jwt v5 | Stateless tokens with `user_id` and `role` claims |
| **Viper** | v1.21.0 | Config management from `.env` files and environment variables |
| **Logrus** | v1.9.4 | Structured, leveled logging |
| **Wire** | v0.7.0 | Compile-time dependency injection |
| **golang-migrate** | latest | Versioned SQL schema migrations |
| **Mockery** | latest | Auto-generated interface mocks for unit testing |
| **Docker Compose** | — | Local development environment orchestration |
| **bcrypt** | x/crypto | Secure password hashing |

### Clean Architecture Layers

```
HTTP Request → Handler → Usecase → Repository → Database / Redis
     ↑                                                   |
     └──────────────── DTO (Response) ───────────────────┘
```

| Layer | Responsibility |
|---|---|
| **Handler** | HTTP concerns: request binding, input validation, status code mapping |
| **Usecase** | Business logic; communicates with other modules via interfaces only |
| **Repository** | Data access abstraction; DB implementation is independently swappable |
| **Entity** | Plain domain structs; no HTTP tags, no business logic, no cross-module fields |

---

## 🔄 Architecture Overview

### JWT Session Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant U as UserUsecase
    participant DB as MySQL
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
    participant PS as ProductService
    participant R as Redis

    Note over C,R: ADD ITEM
    C->>H: POST /api/v1/cart · { product_id, quantity }
    H->>U: AddToCart(userID, productID, quantity)
    U->>PS: GetByProductID(productID)
    PS-->>U: Product{ id, price, stock }
    U->>R: HINCRBY cart:{userID} {productID} {qty}
    U->>R: EXPIRE cart:{userID} 7d
    H-->>C: 200 · item added

    Note over C,R: GET CART — single bulk query, no N+1
    C->>H: GET /api/v1/cart
    H->>U: GetCartDetail(userID)
    U->>R: HGETALL cart:{userID}
    R-->>U: map[ product_id → quantity ]
    U->>PS: GetProductsByIDs([ ids... ]) — one WHERE IN query
    PS-->>U: []Product
    U->>U: compute subtotal + grand_total · detect unavailable_items
    H-->>C: 200 · CartDetailResponse
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
e-commerce/
│
├── Dockerfile                   # Multi-stage build: Go builder → Alpine runtime
├── docker-compose.yml           # App + MySQL 8.0 + Redis 7 in ecommerce-network
├── .dockerignore
├── .mockery.yml                 # Mockery config: registers all interfaces for mock generation
│
├── cmd/
│   └── api/
│       ├── main.go              # Entry point: --seed flag, graceful shutdown
│       ├── application.go       # App struct wiring
│       ├── injector.go          # Wire provider declarations
│       └── wire_gen.go          # Auto-generated DI (do not edit)
│
├── database/
│   ├── migrations/              # Timestamp-prefixed .up.sql / .down.sql pairs
│   │   ├── *_create_users_table.up.sql
│   │   ├── *_create_categories_table.up.sql
│   │   ├── *_create_products_table.up.sql
│   │   ├── *_create_orders_table.up.sql
│   │   └── *_create_order_items_table.up.sql
│   └── seeder/
│       └── seeder.go            # Seed admin user + sample categories
│
├── dependency/                  # Infrastructure adapters
│   ├── gin.go                   # Gin engine setup
│   ├── gorm.go                  # GORM DSN + connection pool (MaxOpen=25, MaxIdle=10)
│   ├── redis.go                 # Redis client + interface (Check/Set/Delete)
│   │                            #   AuthPrefix = "auth:session:"
│   │                            #   CartPrefix = "cart:"  |  CartTTL = 7 days
│   ├── logrus.go                # Structured logger setup
│   ├── validator.go             # go-playground/validator singleton
│   └── viper.go                 # .env → Viper config
│
├── internal/
│   │
│   ├── user/                    # MODULE: Auth & User Identity
│   │   ├── entity/user.go       # User{ ID, Name, Email, Password, Role }
│   │   ├── dto/                 # Request/Response structs with validator tags
│   │   ├── repository/          # FindByEmail, FindByID, Create
│   │   ├── mocks/               # MockUserRepository, MockJwtToken
│   │   └── usecase/
│   │       ├── user_usecase.go       # Interface: Register, Login, GetProfile, Logout
│   │       ├── user_usecase_impl.go  # bcrypt, JWT creation, Redis session lifecycle
│   │       └── user_usecase_test.go  # 4 test functions (Login, Register, Logout, GetProfile)
│   │
│   ├── product/                 # MODULE: Catalog
│   │   ├── entity/
│   │   │   ├── category.go      # Category{ ID, Name, Slug }
│   │   │   └── product.go       # Product{ ID, CategoryID, Name, Slug, Price, Stock, SKU, IsActive }
│   │   │                        #   ProductFilter for paginated queries
│   │   ├── dto/                 # Create/Update/Search/StockAdjust requests; paginated response
│   │   ├── repository/          # CRUD + FindByIDs, FindAll(filter), DecreaseStock, AdjustStock
│   │   ├── mocks/               # MockProductRepository, MockCategoryRepository,
│   │   │                        # MockCategoryUsecase, MockProductService
│   │   └── usecase/
│   │       ├── category_usecase.go        # Interface: CreateCategory, GetAllCategories, ValidateCategoryExists
│   │       ├── category_usecase_impl.go   # Slug generation, duplicate check
│   │       ├── category_usecase_test.go   # 3 test functions
│   │       ├── product_usecase.go         # Interface: Create, Update, Delete, Search, GetDetail, AdjustStock
│   │       ├── product_usecase_impl.go    # Full lifecycle; re-validates category only on change
│   │       ├── product_usecase_test.go    # 9 test functions (full CRUD + service contract methods)
│   │       └── contract.go                # ProductService{ GetByProductID, GetProductsByIDs, DecreaseStock }
│   │
│   ├── cart/                    # MODULE: Shopping Cart (Redis-backed)
│   │   ├── dto/
│   │   │   ├── cart_request.go   # CartItemCreateRequest, CartItemUpdateRequest
│   │   │   └── cart_response.go  # CartDetailResponse{ Items, UnavailableItems, GrandTotal }
│   │   ├── repository/           # Interface + Redis impl: HINCRBY/HSET/HDEL/HGETALL/DEL
│   │   └── usecase/
│   │       ├── cart_usecase.go        # Interface: AddToCart, UpdateCartItem, RemoveFromCart, GetCartDetail
│   │       ├── cart_usecase_impl.go   # Bulk enrichment, unavailable_items detection, grand_total
│   │       └── contract.go            # CartService{ GetRawCart, ClearCart }
│   │
│   ├── middleware/
│   │   ├── auth_middleware.go    # Token extraction → JWT validation → Redis session check
│   │   ├── admin_middleware.go   # Role check from JWT context; 403 if not admin
│   │   └── logger_middleware.go  # Per-request: method, path, status, latency_ms, client_ip
│   │
│   └── mocks/                   # Shared mocks: MockRedis, MockTransaction
│
├── pkg/
│   ├── response/response.go     # ResponseSuccess / ResponseError envelope helpers
│   ├── jwt/jwt.go               # JwtToken interface + HS256 impl; Auth{ UserID, Role }
│   ├── apperror/apperror.go     # Typed sentinel errors (ErrProductNotFound, ErrInsufficientStock, etc.)
│   ├── skugen/                  # SKU auto-generation
│   └── transaction/             # WithTransaction(ctx, fn) — isolates TX boilerplate from usecases
│
└── routes/router.go             # Route groups: public · protected · admin
```

### Module Boundary Rules

```
✅ ALLOWED:   internal/cart   → internal/product/usecase  (via contract.go interface)
✅ ALLOWED:   internal/order  → internal/cart/usecase     (via contract.go interface)
❌ FORBIDDEN: internal/cart   → internal/product/entity   (direct struct dependency)
❌ FORBIDDEN: internal/cart   → internal/product/repository (direct data access)
```

Each module exposes its capabilities to other modules **only** through a `contract.go` file containing Go interfaces — a clear and auditable boundary similar to an API contract between services.

---

## 📡 API Endpoints

### Authentication & User

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/register` | Register a new user | — |
| `POST` | `/api/v1/login` | Login and receive JWT session token | — |
| `GET` | `/api/v1/profile` | Get authenticated user profile | ✅ JWT |
| `DELETE` | `/api/v1/logout` | Revoke current JWT session | ✅ JWT |

### Product Catalog

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api/v1/categories` | List all categories | — |
| `POST` | `/api/v1/admin/categories` | Create category (auto slug) | ✅ Admin |
| `GET` | `/api/v1/products` | List & search products with pagination | — |
| `GET` | `/api/v1/products/:product_id` | Get product detail | — |
| `POST` | `/api/v1/admin/products` | Create product | ✅ Admin |
| `PUT` | `/api/v1/admin/products/:product_id` | Update product | ✅ Admin |
| `DELETE` | `/api/v1/admin/products/:product_id` | Soft-delete product | ✅ Admin |
| `PATCH` | `/api/v1/admin/products/:product_id/adjust-stock` | Set stock value | ✅ Admin |

### Cart

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api/v1/cart` | View cart with enriched product data | ✅ JWT |
| `POST` | `/api/v1/cart` | Add item to cart | ✅ JWT |
| `PATCH` | `/api/v1/cart/:product_id` | Update item quantity | ✅ JWT |
| `DELETE` | `/api/v1/cart/:product_id` | Remove single item | ✅ JWT |
| `DELETE` | `/api/v1/cart` | Clear entire cart | ✅ JWT |

### Checkout & Orders *(Planned)*

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/checkout` | Create order from cart | ✅ JWT |
| `GET` | `/api/v1/orders` | List user orders | ✅ JWT |
| `GET` | `/api/v1/orders/:id` | Order detail | ✅ JWT |

---

## 📦 Request & Response Samples

### `POST /api/v1/register`

```json
// Request
{ "name": "Budi Santoso", "email": "budi@example.com", "password": "securepassword123" }

// 201 Created
{ "success": true, "message": "user registered successfully", "data": { "id": 1, "name": "Budi Santoso", "email": "budi@example.com" } }

// 409 Conflict
{ "success": false, "message": "Email Already Exists" }
```

### `POST /api/v1/login`

```json
// Request
{ "email": "budi@example.com", "password": "securepassword123" }

// 200 OK
{ "success": true, "message": "user logged in successfully", "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." } }

// 401 Unauthorized
{ "success": false, "message": "Wrong Email or Password" }
```

### `GET /api/v1/products?search=keyboard&category_id=1&page=1&limit=10`

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

### `POST /api/v1/admin/products`

```json
// Request — Authorization: Bearer <admin_token>
{
  "category_id": 1, "name": "Mechanical Keyboard 60%",
  "description": "Hot-swappable switches, RGB backlight", "price": 750000, "stock": 50
}

// 201 Created
{
  "success": true, "message": "Product created successfully",
  "data": { "id": 12, "slug": "mechanical-keyboard-60", "sku": "PRD-XXXXXXXX", "is_active": true, ... }
}

// 404 — category not found
{ "success": false, "message": "Category Not Found" }
```

### `POST /api/v1/cart` & `GET /api/v1/cart`

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

> **`unavailable_items`**: Products that exist in Redis but have since been deactivated or deleted from the catalog. They are surfaced for user review rather than silently removed.

### `DELETE /api/v1/logout`

```json
// 200 OK
{ "success": true, "message": "user logged out successfully" }
// Token is immediately invalid — Redis key deleted.
```

---

## 🚀 Getting Started

You can run this project using **Docker Compose** (recommended) or manually.

---

### Option A: Docker Compose (Recommended)

This starts the API, MySQL, and Redis together in an isolated network. No local database setup required.

**1. Clone the repository**
```bash
git clone https://github.com/Mpayy/e-commerce.git
cd e-commerce
```

**2. Configure environment variables**
```bash
cp .env.example .env
# Edit .env with your preferred values
```

**3. Start all services**
```bash
docker compose up --build -d
```

**4. Run database migrations** (inside the running container)
```bash
docker exec -it ecommerce-app sh -c \
  "migrate -database 'mysql://root:root@tcp(mysql:3306)/ecommerce' -path database/migrations up"
```

> The `DATABASE_NAME` for Docker Compose is `ecommerce` (set in `docker-compose.yml`).

**5. (Optional) Seed initial data**
```bash
docker exec -it ecommerce-app ./main --seed
```

**6. Verify the server is running**
```bash
curl -s http://localhost:8080/api/v1/products | jq .
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
| MySQL | 8.0+ |
| Redis | 7.0+ |
| `golang-migrate` CLI | latest |

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
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
DATABASE_HOST=127.0.0.1
DATABASE_PORT=3306
DATABASE_NAME=ecommerce_db
DATABASE_USERNAME=root
DATABASE_PASSWORD=your_password

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_DB=0

APP_HOST=0.0.0.0
APP_PORT=8080

LOG_LEVEL=debug

# Use a long, random string in production
JWT_SECRET_KEY=change-me-to-a-strong-random-secret-key
```

**3. Create database**
```bash
mysql -u root -p -e "CREATE DATABASE ecommerce_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

**4. Run migrations**
```bash
migrate -database "mysql://root:your_password@tcp(127.0.0.1:3306)/ecommerce_db" \
        -path database/migrations up
```

To rollback:
```bash
migrate -database "mysql://root:your_password@tcp(127.0.0.1:3306)/ecommerce_db" \
        -path database/migrations down 1
```

**5. (Optional) Seed initial data**
```bash
go run ./cmd/api/... --seed
```

**6. Run the application**
```bash
go run ./cmd/api/...
```

Server starts at `http://0.0.0.0:8080`.

---

### Developer Commands

**Regenerate Wire DI** (after modifying `injector.go`):
```bash
cd cmd/api && go generate ./...
```

**Regenerate mocks** (after modifying any interface):
```bash
go generate ./...
```

**Run unit tests:**
```bash
go test ./internal/... -v -race
```

The `-race` flag enables Go's built-in data race detector. All business logic tests run without a live database or Redis — dependencies are mocked via `mockery`.

**Run tests for a specific module:**
```bash
# User module
go test -v ./internal/user/usecase -run "TestUserUsecaseImpl"

# Product module
go test -v ./internal/product/usecase -run "TestProductUsecaseImpl"

# Category module
go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl"
```

---

## 📝 Architecture Notes

### No Foreign Key Constraints Between Modules

Each module (`user`, `product`, `cart`, `order`) is treated as an independent **Bounded Context**. Cross-module database-level FK constraints would create infrastructure-layer coupling that conflicts with the modular design goal.

Instead:
- `products.category_id` is a plain `uint` field — GORM will not auto-create an FK constraint without an explicit struct association field
- `order_items` stores a snapshot of `product_name` and `price` at checkout time, making order history immutable regardless of future catalog changes

If the need arises to split modules into independent services, the only change required would be replacing Go interface calls with gRPC or HTTP calls. The database schema needs no restructuring.

### Cart Enrichment Without N+1

The `GET /cart` flow uses a two-step approach:

1. **One Redis call**: `HGETALL cart:{userID}` returns the full `map[product_id → quantity]`
2. **One DB query**: `FindByIDs` executes a single `WHERE id IN (...)` with all product IDs from step 1
3. Everything else (subtotal, grand_total, unavailable_items) is computed in memory

This keeps database round trips constant regardless of cart size.

### Concurrent Stock Deduction (Checkout, Planned)

`ProductRepository.DecreaseStock` already implements pessimistic row-level locking:

```go
tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID)
```

When two concurrent checkout requests compete for the same low-stock product, only one goroutine acquires the lock. The other waits, then fails the stock check and rolls back cleanly. The Redis cart is preserved on rollback so the user can retry without re-adding items.

### Transaction Abstraction

`pkg/transaction.WithTransaction(ctx, fn)` wraps transaction boilerplate, keeping it out of usecase logic. This also allows the transaction itself to be mocked in unit tests — usecases remain testable without a live database connection.

### Connection Pool Configuration

```
MaxOpenConns:    25    — max simultaneous connections
MaxIdleConns:    10    — connections kept ready in the pool
ConnMaxLifetime: 5min  — recycles connections to prevent stale handles
ConnMaxIdleTime: 1min  — evicts idle connections sooner
```

---

<div align="center">
<sub>MIT License · Contributions welcome</sub>
</div>
