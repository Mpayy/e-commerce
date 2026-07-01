<div align="center">

# 🛒 E-Commerce API

### Go · Gin · GORM · MySQL · Redis · JWT

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12.0-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://github.com/gin-gonic/gin)
[![GORM](https://img.shields.io/badge/GORM-v1.31.2-E10098?style=for-the-badge)](https://gorm.io/)
[![Redis](https://img.shields.io/badge/Redis-v9.21.0-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=for-the-badge)]()

<p align="center">
  <em>A production-grade, modular monolith e-commerce backend built with <strong>Clean Architecture</strong> principles and deliberate engineering decisions — designed to scale horizontally into microservices when business demands it.</em>
</p>

<p align="center">
  Implements <strong>session-based token management via Redis</strong> for instant token revocation on logout — a pattern used in high-traffic production systems where JWT invalidation at the application layer is both faster and more reliable than database polling.
</p>

</div>

---

## 📋 Table of Contents

- [Key Features & Roadmap](#-key-features--roadmap)
- [Tech Stack & Architectural Decisions](#-tech-stack--architectural-decisions)
- [System Architecture](#-system-architecture)
- [Directory Structure](#-directory-structure)
- [API Endpoints](#-api-endpoints)
- [Request & Response Samples](#-request--response-samples)
- [Local Setup Guide](#-local-setup-guide)
- [Engineering Notes](#-engineering-notes)

---

## ✅ Key Features & Roadmap

> Features are scoped into **Implemented** (current codebase) and **Planned** (defined in the architecture blueprint).

### 🟢 Implemented

- [x] **User Registration** — bcrypt password hashing with configurable cost factor; duplicate email detection with domain-specific error typing
- [x] **JWT Authentication (HS256)** — 30-day token lifecycle; signed with a configurable secret key loaded via Viper
- [x] **Redis Session Management (Token Revocation)** — On login, JWT is stored in Redis under `auth:session:<token>`; on logout, the key is atomically deleted — making the token immediately invalid even before JWT expiry
- [x] **Auth Middleware** — Parses `Authorization: Bearer <token>` → validates JWT signature → checks token existence in Redis → injects `auth` and `token` into Gin context
- [x] **Role-Based Access Control (RBAC)** — `AdminMiddleware` inspects the `role` claim from the JWT payload; returns `403 Forbidden` for non-admin routes
- [x] **User Profile** — Authenticated endpoint returning user data from a JWT-derived `user_id` lookup
- [x] **Category Management (Admin)** — Create + list categories with slug auto-generation; conflict detection on duplicate names
- [x] **Product Management (Admin)** — Create product with SKU auto-generation, category validation, and duplicate SKU/name detection
- [x] **Structured Logging** — Logrus-powered request logger middleware with per-request `method`, `path`, `status`, `latency_ms`, and `client_ip` fields; log level adapts to HTTP status range (Info / Warn / Error)
- [x] **Standardized JSON Response** — Global `ResponseSuccess` / `ResponseError` helpers enforce a consistent response envelope across all handlers
- [x] **Versioned SQL Migrations** — `golang-migrate`-compatible timestamp-prefixed SQL files; intentionally **no FOREIGN KEY constraints** between modules (by design — see [Engineering Notes](#-engineering-notes))
- [x] **Graceful Shutdown** — `SIGINT`/`SIGTERM` signal handling with a 5-second drain window before forced shutdown
- [x] **Dependency Injection via Wire** — Google Wire generates compile-time DI wiring; zero runtime reflection overhead
- [x] **Database Seeder** — Bootstrap admin/customer seed data via `--seed` CLI flag
- [x] **Unit Tests with Mocks** — `mockery`-generated mocks for `Redis` and `JwtToken` interfaces; test coverage for `UserUsecase` (register, login, logout, profile)

### 🔜 Planned (Blueprint Defined)

- [ ] **Product Listing & Search** — `GET /products` with `search`, `category_id`, `page`, and `limit` query params; paginated response with `meta`
- [ ] **Product Detail** — `GET /products/:id`
- [ ] **Product Stock Adjustment** — `PATCH /admin/products/:id/stock` for restock operations with reason tracking
- [ ] **Product Update & Delete** — `PUT`/`DELETE` admin endpoints
- [ ] **Cart System (Redis Hash)** — Cart stored as `HSET cart:{user_id} {product_id} {quantity}`; 7-day TTL auto-refreshed on every write; price never stored in cart (always fetched real-time from Product module via interface call)
- [ ] **Cart Enrichment** — `GET /cart` fetches raw quantities from Redis, then makes a single bulk `productService.GetByIDs()` call to enrich with names, prices, and available stock — explicitly avoids N+1
- [ ] **Checkout (ACID-safe)** — Converts Redis cart to MySQL `orders`+`order_items` inside a single DB transaction; uses `SELECT ... FOR UPDATE` row-level locking via `productService.DecreaseStock(tx, ...)` to prevent race-condition oversell
- [ ] **Order History** — `GET /orders` and `GET /orders/:id` for authenticated users
- [ ] **Invoice Snapshot** — `OrderItem` stores `product_name` and `price` at transaction time — history is immutable regardless of future product edits
- [ ] **Saga Pattern Readiness** — Checkout rollback leaves Redis cart intact; user can safely retry without re-adding items
- [ ] **Swagger / OpenAPI Docs** — `swaggo`-generated interactive API documentation
- [ ] **Docker Compose** — One-command local environment spin-up (App + MySQL + Redis)

---

## 🏗️ Tech Stack & Architectural Decisions

| Technology | Version | Engineering Justification |
|---|---|---|
| **Go** | 1.26.4 | Compiled, garbage-collected, goroutine-based concurrency — delivers sub-millisecond p99 latency for I/O-bound API workloads with minimal operational overhead |
| **Gin** | v1.12.0 | Zero-allocation router with radix-tree path matching; Sonic JSON encoder under the hood — benchmarks significantly faster than `net/http` with `encoding/json` on serialization-heavy paths |
| **GORM** | v1.31.2 | ORM with explicit query control; `ParameterizedQueries: true` enforces SQL injection prevention at the driver level; custom `logrusWriter` integrates slow-query logging |
| **MySQL** | 8.0+ | ACID-compliant relational store with `SELECT ... FOR UPDATE` row-level locking — prerequisite for the concurrent checkout stock-deduction pattern |
| **Redis** | go-redis v9 | Session store for JWT tokens (`auth:session:<token>`); `DEL` on logout provides O(1) immediate revocation without scanning or blacklist tables; also serves as the Cart storage layer (`HSET`/`HGETALL`) |
| **JWT (HS256)** | golang-jwt v5 | Stateless token format with `user_id` and `role` claims embedded; Redis existence check converts it into a stateful session without sacrificing bearer-token ergonomics |
| **Viper** | v1.21.0 | Hierarchical config with `.env` file support and environment-variable override — 12-factor app compliant |
| **Logrus** | v1.9.4 | Structured, leveled logging with JSON output capability; request middleware logs `latency_ms` per request enabling p95/p99 monitoring integration |
| **Wire** | v0.7.0 | Compile-time dependency injection — injection errors surface at `go generate`, not at runtime; zero reflection overhead in the hot path |
| **golang-migrate** | latest | Versioned, Git-reviewable SQL migrations; explicit `CREATE TABLE` without `FOREIGN KEY` makes cross-module data boundaries unambiguous and schema evolution auditable |
| **Mockery** | latest | Interface-driven mocks auto-generated from `//go:generate` annotations; enables isolated unit testing of business logic without database or Redis dependencies |
| **bcrypt** | x/crypto | Password hashing with `DefaultCost` (10 rounds); resistant to timing attacks; cost factor is tunable for future hardware scaling |

### Clean Architecture Layer Contracts

```
HTTP Request --> Handler --> Usecase --> Repository --> Database/Redis
     ^                                                       |
     └──────────────── DTO (Response) ──────────────────────┘
```

- **Handler** layer: HTTP protocol concerns only — binding, validation, status code mapping
- **Usecase** layer: Pure business logic, orchestrates cross-domain calls via **interfaces only**
- **Repository** layer: Data access abstraction; database implementation is fully swappable
- **Entity** layer: Naked domain structs — no HTTP tags, no business logic, no cross-module associations

---

## 🔄 System Architecture

### Session-Based JWT Flow (Login & Logout)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as UserHandler
    participant U as UserUsecase
    participant DB as MySQL
    participant J as pkg/jwt
    participant R as Redis

    Note over C,R: LOGIN FLOW
    C->>H: POST /api/v1/login · { email, password }
    H->>H: Validate request struct (go-playground/validator)
    H->>U: Login(ctx, request)
    U->>DB: FindByEmail(email)
    DB-->>U: User{id, email, password_hash, role}
    U->>U: bcrypt.CompareHashAndPassword()
    U->>J: CreateToken({user_id, role})
    J-->>U: signed JWT string (HS256, 30d TTL)
    U->>R: SET auth:session:TOKEN · {user_id,role} · EX 2592000
    R-->>U: OK
    U-->>H: TokenResponse{ token }
    H-->>C: 200 OK · { success: true, data: { token } }

    Note over C,R: AUTHENTICATED REQUEST FLOW
    C->>H: GET /api/v1/profile · Authorization: Bearer TOKEN
    H->>H: AuthMiddleware - Extract Bearer token
    H->>J: ParseToken(token) - validate signature + expiry
    J-->>H: Auth{ user_id, role }
    H->>R: EXISTS auth:session:TOKEN
    R-->>H: 1 (key exists)
    H->>H: ctx.Set("auth", Auth) · ctx.Set("token", token)
    H->>H: ctx.Next() - route handler proceeds
    H-->>C: 200 OK · profile data

    Note over C,R: LOGOUT FLOW
    C->>H: DELETE /api/v1/logout · Authorization: Bearer TOKEN
    H->>H: AuthMiddleware passes (token still valid in Redis)
    H->>U: Logout(ctx, token)
    U->>R: DEL auth:session:TOKEN
    R-->>U: 1 (key deleted)
    U-->>H: nil error
    H-->>C: 200 OK · { success: true }
    Note right of R: Token permanently invalid. Any subsequent request returns 401.
```

### Token Revocation Strategy Comparison

| Approach | Lookup Cost | Scalability | Instant Revocation |
|---|---|---|---|
| Pure Stateless JWT | O(0) — no lookup | ✅ Infinite | ❌ Must wait for expiry |
| DB Blacklist Table | O(log n) — indexed query | ⚠️ DB bottleneck under load | ✅ Yes |
| **Redis Session Store (this project)** | **O(1) — hash lookup** | **✅ Redis cluster-ready** | **✅ Immediate on DEL** |

---

## 📁 Directory Structure

```
e-commerce/
│
├── cmd/
│   └── api/
│       ├── main.go              # Entry point: CLI flags (--seed), graceful shutdown, server lifecycle
│       ├── application.go       # Application struct binding (Gin, DB, Config, Logger)
│       ├── injector.go          # Wire provider declarations
│       └── wire_gen.go          # Auto-generated DI wiring (do not edit manually)
│
├── database/
│   ├── migrations/              # Versioned SQL files (timestamp-prefixed, .up/.down pairs)
│   │   ├── *_create_users_table.up.sql
│   │   ├── *_create_categories_table.up.sql
│   │   ├── *_create_products_table.up.sql
│   │   ├── *_create_orders_table.up.sql
│   │   └── *_create_order_items_table.up.sql
│   └── seeder/
│       └── seeder.go            # Bootstrap seed data (admin user, sample categories)
│
├── dependency/                  # Infrastructure adapters for external services
│   ├── app.go                   # Application container struct
│   ├── gin.go                   # Gin engine initialization (mode, trusted proxies)
│   ├── gorm.go                  # GORM setup: DSN, connection pool (MaxOpen=25, MaxIdle=10)
│   ├── redis.go                 # Redis client + Redis interface (CheckToRedis, SetToRedis, DeleteFromRedis)
│   ├── logrus.go                # Logger initialization with configurable log level
│   ├── validator.go             # go-playground/validator singleton
│   └── viper.go                 # Config loader: reads .env -> Viper struct
│
├── internal/                    # All business domain code — the core of the application
│   │
│   ├── user/                    # MODULE: Auth & User Identity
│   │   ├── entity/
│   │   │   └── user.go          # Domain model: User{ID, Name, Email, Password, Role}
│   │   │                        #   Role constants: "customer" | "admin"
│   │   ├── dto/
│   │   │   ├── user_request.go  # UserRegisterRequest, UserLoginRequest (with validator tags)
│   │   │   └── user_response.go # UserResponse (no password field), TokenResponse{Token}
│   │   ├── repository/          # Data access: FindByEmail, FindByID, Create (within TX)
│   │   ├── usecase/
│   │   │   ├── user_usecase.go       # Interface: Register, Login, GetProfile, Logout
│   │   │   ├── user_usecase_impl.go  # Impl: bcrypt hashing, JWT creation, Redis session write/delete
│   │   │   └── user_usecase_test.go  # Unit tests with mockery-generated Redis + JWT mocks
│   │   ├── delivery/http/
│   │   │   ├── user_handler.go       # Interface: Register, Login, GetProfile, Logout
│   │   │   └── user_handler_impl.go  # HTTP adapter: bind JSON → validate → usecase → respond
│   │   └── mocks/                # Auto-generated mocks for JWT and Redis interfaces
│   │
│   ├── product/                 # MODULE: Catalog (Products & Categories)
│   │   ├── entity/
│   │   │   ├── category.go      # Category{ID, Name, Slug, CreatedAt, UpdatedAt}
│   │   │   └── product.go       # Product{ID, CategoryID, Name, Slug, Price, Stock, SKU, IsActive}
│   │   │                        #   CategoryID is a plain uint — NOT a GORM association field
│   │   │                        #   => GORM will never auto-generate a FOREIGN KEY constraint
│   │   ├── dto/
│   │   │   ├── category_request.go   # CategoryRequest{Name} with validator tags
│   │   │   ├── category_response.go  # CategoryResponse DTO
│   │   │   ├── product_request.go    # ProductRequest{CategoryID, Name, Description, Price, Stock, SKU}
│   │   │   └── product_response.go   # ProductResponse (full fields including auto-generated slug)
│   │   ├── repository/          # Category: FindAll, Create, FindByID | Product: Create, FindByID
│   │   ├── usecase/
│   │   │   ├── category_usecase.go       # Interface: CreateCategory, GetAllCategories
│   │   │   ├── category_usecase_impl.go  # Slug generation (gosimple/slug), duplicate check
│   │   │   ├── product_usecase.go        # Interface: CreateProduct
│   │   │   └── product_usecase_impl.go   # Category validation, SKU auto-generation (pkg/skugen)
│   │   └── delivery/http/
│   │       ├── category_handler.go       # Interface: Create, GetAll
│   │       ├── category_handler_impl.go  # Handles 409 Conflict on duplicate category name
│   │       ├── product_handler.go        # Interface: Create
│   │       └── product_handler_impl.go   # Handles 404 (category not found), 409 (duplicate SKU/name)
│   │
│   ├── middleware/              # CROSS-CUTTING: HTTP Middleware
│   │   ├── auth_middleware.go   # 1. Extract Bearer token
│   │   │                        # 2. ParseToken (JWT signature + expiry validation)
│   │   │                        # 3. EXISTS auth:session:<token> in Redis
│   │   │                        # 4. ctx.Set("auth", Auth) -> ctx.Next()
│   │   ├── admin_middleware.go  # Reads auth.Role from context; 403 if role != "admin"
│   │   └── logger_middleware.go # Per-request structured log: method, path, status, latency_ms, client_ip
│   │
│   └── mocks/                   # Shared mocks (mock_redis.go — used across module tests)
│
├── pkg/                         # Shared utilities — zero domain-specific logic
│   ├── response/
│   │   └── response.go          # ResponseSuccess(ctx, code, message, data)
│   │                            # ResponseError(ctx, code, message, err) -> AbortWithStatusJSON
│   ├── jwt/
│   │   └── jwt.go               # JwtToken interface + HS256 implementation
│   │                            # Auth{UserID, Role} struct | TokenDuration = 30 days
│   ├── apperror/
│   │   └── apperror.go          # Sentinel errors: ErrNotFound, ErrUnauthorized, ErrDuplicatedEmail...
│   │                            # ExtractValidationErrors() maps FieldError -> map[field]message
│   ├── skugen/                  # Auto-generates SKU string if not provided in ProductRequest
│   └── transaction/
│       └── transaction.go       # WithTransaction(ctx, fn) wrapper — isolates TX boilerplate from usecases
│
└── routes/
    └── router.go                # Route groups: public | protected (AuthMiddleware) | admin (+AdminMiddleware)
```

### Module Isolation Rules (Enforced at Code Level)

```
ALLOWED:   internal/order  -> imports internal/product/usecase  (interface contract only)
FORBIDDEN: internal/order  -> imports internal/product/entity   (direct struct coupling)
FORBIDDEN: internal/order  -> imports internal/product/repository (direct DB access)
```

Every module that exposes capabilities to other modules **must** do so exclusively through a `contract.go` file containing only Go interfaces. This is the single, auditable boundary between modules — directly analogous to a service API surface in a microservices architecture.

---

## 📡 API Endpoints

### Authentication & User

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/register` | Register a new user account | No |
| `POST` | `/api/v1/login` | Authenticate and receive a JWT session token | No |
| `GET` | `/api/v1/profile` | Retrieve the authenticated user's profile | ✅ JWT |
| `DELETE` | `/api/v1/logout` | Revoke the current JWT session token in Redis | ✅ JWT |

### Product Catalog

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/categories` | List all product categories | No |
| `POST` | `/api/v1/admin/categories` | Create a new category (slug auto-generated) | ✅ JWT + Admin |
| `GET` | `/api/v1/products` *(planned)* | Search & list products with pagination | No |
| `GET` | `/api/v1/products/:id` *(planned)* | Get product detail by ID | No |
| `POST` | `/api/v1/admin/products` | Create a new product (validates category, auto-generates SKU) | ✅ JWT + Admin |
| `PUT` | `/api/v1/admin/products/:id` *(planned)* | Update product details | ✅ JWT + Admin |
| `DELETE` | `/api/v1/admin/products/:id` *(planned)* | Soft-delete a product | ✅ JWT + Admin |
| `PATCH` | `/api/v1/admin/products/:id/stock` *(planned)* | Manual stock adjustment (restock) | ✅ JWT + Admin |

### Cart & Checkout *(Planned)*

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/cart` | View cart items (enriched with live product data) | ✅ JWT |
| `POST` | `/api/v1/cart/items` | Add a product to cart | ✅ JWT |
| `PUT` | `/api/v1/cart/items/:product_id` | Update item quantity | ✅ JWT |
| `DELETE` | `/api/v1/cart/items/:product_id` | Remove a single item from cart | ✅ JWT |
| `DELETE` | `/api/v1/cart` | Clear entire cart | ✅ JWT |
| `POST` | `/api/v1/checkout` | Convert Redis cart to MySQL order (ACID transaction) | ✅ JWT |
| `GET` | `/api/v1/orders` | List current user's order history | ✅ JWT |
| `GET` | `/api/v1/orders/:id` | Get order detail / invoice | ✅ JWT |

---

## 📦 Request & Response Samples

### `POST /api/v1/register`

**Request Body:**
```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "password": "securepassword123"
}
```

**201 Created:**
```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "email": "budi@example.com"
  }
}
```

**400 Bad Request (validation failure):**
```json
{
  "success": false,
  "message": "validation failed",
  "error": {
    "email": "email must be a valid email address",
    "password": "password must be at least 6 characters"
  }
}
```

**409 Conflict (duplicate email):**
```json
{
  "success": false,
  "message": "email already registered",
  "error": null
}
```

---

### `POST /api/v1/login`

**Request Body:**
```json
{
  "email": "budi@example.com",
  "password": "securepassword123"
}
```

**200 OK:**
```json
{
  "success": true,
  "message": "user logged in successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk2NjM..."
  }
}
```

**401 Unauthorized (wrong credentials):**
```json
{
  "success": false,
  "message": "wrong email or password",
  "error": null
}
```

---

### `POST /api/v1/admin/products`

> Requires header: `Authorization: Bearer <admin_token>`

**Request Body:**
```json
{
  "category_id": 1,
  "name": "Mechanical Keyboard 60%",
  "description": "Hot-swappable switches, RGB backlight, compact 60% layout",
  "price": 750000,
  "stock": 50,
  "sku": "KBD-60-001"
}
```

**201 Created:**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": 12,
    "category_id": 1,
    "name": "Mechanical Keyboard 60%",
    "slug": "mechanical-keyboard-60",
    "description": "Hot-swappable switches, RGB backlight, compact 60% layout",
    "price": 750000,
    "stock": 50,
    "sku": "KBD-60-001",
    "is_active": true
  }
}
```

**404 Not Found (invalid category):**
```json
{
  "success": false,
  "message": "category not found",
  "error": null
}
```

**403 Forbidden (non-admin token):**
```json
{
  "success": false,
  "message": "forbidden: admin access required",
  "error": null
}
```

---

### `DELETE /api/v1/logout`

> Requires header: `Authorization: Bearer <token>`

**200 OK:**
```json
{
  "success": true,
  "message": "user logged out successfully"
}
```

> After this response, any subsequent request using the same token will immediately return `401 Unauthorized` — the Redis key `auth:session:<token>` has been atomically deleted.

**401 Unauthorized (token already expired or invalid):**
```json
{
  "success": false,
  "message": "unauthorized",
  "error": null
}
```

---

## 🚀 Local Setup Guide

### Prerequisites

| Tool | Minimum Version | Notes |
|---|---|---|
| Go | 1.22+ | [Download](https://go.dev/dl/) |
| MySQL | 8.0+ | Local instance or Docker |
| Redis | 7.0+ | Local instance or Docker |
| `golang-migrate` CLI | latest | For running SQL migrations |

Install `golang-migrate`:
```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

### 1. Clone the Repository

```bash
git clone https://github.com/Mpayy/e-commerce.git
cd e-commerce
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

`.env` reference:
```env
# Database (MySQL)
DATABASE_HOST=127.0.0.1
DATABASE_PORT=3306
DATABASE_NAME=ecommerce_db
DATABASE_USERNAME=root
DATABASE_PASSWORD=your_password

# Redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_DB=0

# Application
APP_HOST=0.0.0.0
APP_PORT=8080

# Logging level: trace | debug | info | warn | error
LOG_LEVEL=debug

# JWT — use a cryptographically random string in production (min 32 chars)
JWT_SECRET_KEY=change-me-to-a-strong-random-secret-key
```

### 4. Create the Database

```bash
mysql -u root -p -e "CREATE DATABASE ecommerce_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 5. Run Database Migrations

```bash
migrate -database "mysql://root:your_password@tcp(127.0.0.1:3306)/ecommerce_db" \
        -path database/migrations \
        up
```

To roll back the last migration batch:
```bash
migrate -database "mysql://root:your_password@tcp(127.0.0.1:3306)/ecommerce_db" \
        -path database/migrations \
        down 1
```

### 6. (Optional) Seed Initial Data

```bash
go run ./cmd/api/... --seed
```

> Seeds an admin user and sample categories to bootstrap development. Run once after migrations.

### 7. Run the Application

```bash
go run ./cmd/api/...
```

The server starts at `http://0.0.0.0:8080` (or your configured `APP_HOST:APP_PORT`).

**Verify the server is running:**
```bash
curl -s http://localhost:8080/api/v1/categories | jq .
```

---

### 8. Regenerate Dependency Injection

Run this after adding or modifying provider functions in `injector.go`:

```bash
cd cmd/api && go generate ./...
```

> Regenerates `wire_gen.go`. Wire injection errors are caught at compile time, not at runtime.

### 9. Regenerate Mocks

Run this after modifying any interface that has `//go:generate mockery` annotations:

```bash
go generate ./...
```

### 10. Run Unit Tests

```bash
go test ./internal/... -v -race
```

> The `-race` flag enables Go's built-in data race detector — essential for validating concurrent usecase logic.

---

## 🔬 Engineering Notes

### Why No FOREIGN KEY Constraints Between Modules?

This is a deliberate architectural decision, not an oversight.

In a Modular Monolith, each module (`user`, `product`, `order`) represents an independent **Bounded Context**. Database-level FK constraints between modules create tight coupling at the infrastructure layer, directly conflicting with the modular design goal:

- `products.category_id` is a **logical reference** (plain `uint`), not a GORM association field. GORM only auto-generates FK constraints when you define a struct association (e.g., `Category Category`). Since we deliberately omit that field, no FK constraint is ever created — confirmed by reviewing every entity struct in the codebase.
- `order_items.product_id` is a **cross-module logical reference** — `OrderItem` will never `JOIN` against `products`. Instead, it stores a **snapshot** of `product_name` and `price` at checkout time, making invoices historically immutable regardless of future catalog changes.

**Path to Microservices:** When the time comes to split modules into separate services with independent databases, the only required change is replacing Go interface calls (e.g., `productService.GetByID()`) with gRPC/HTTP calls. The database schema requires zero restructuring.

### Checkout Race Condition Prevention

The planned checkout flow uses `SELECT ... FOR UPDATE` (row-level pessimistic locking) via GORM's `clause.Locking` within a single MySQL transaction:

```go
tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID)
```

When two concurrent checkout requests compete for the same product with `stock = 1`, only one goroutine acquires the row lock. The second waits, then fails the stock check and rolls back the entire transaction. The Redis cart is **not deleted** on rollback — the user can retry seamlessly without re-adding items.

### Transaction Abstraction (`pkg/transaction`)

Database transaction management is abstracted behind a `WithTransaction(ctx, fn)` interface. This prevents transaction boilerplate from leaking into usecase logic and keeps usecases unit-testable without a live database connection — the transaction itself can be mocked in tests.

### Connection Pool Tuning

The GORM setup in `dependency/gorm.go` is pre-configured for production workloads:

```
MaxOpenConns:    25       # Max simultaneous DB connections
MaxIdleConns:    10       # Connections held ready in pool
ConnMaxLifetime: 5 min    # Recycles connections to avoid stale handles
ConnMaxIdleTime: 1 min    # Evicts idle connections faster
```

These values align with common production tuning for a single-instance API serving moderate traffic loads.

---

<div align="center">

**Built with deliberate engineering decisions, not tutorial defaults.**

</div>
