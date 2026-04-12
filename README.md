# 🏎️ Pistoni Core

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)
![Framework](https://img.shields.io/badge/Framework-Gin-0081CB?style=flat-square)
![Database](https://img.shields.io/badge/Database-Prisma-2D3748?style=flat-square&logo=prisma)
![Realtime](https://img.shields.io/badge/Realtime-Socket.io-010101?style=flat-square&logo=socket.io)

**Pistoni Core** is the high-performance Go backend designed for the incremental migration of the Pistoni ecosystem from NestJS. Focused on reliability, speed, and clean architecture, it serves as the new backbone for the Pistoni API.

---

## 🚀 Vision: Incremental Migration

This project implements a strategic, piece-by-piece migration from a legacy NestJS architecture to Go. By maintaining endpoint parity and sharing the same database (via Prisma), we ensure a seamless transition while drastically improving resource efficiency and execution speed.

## 🛠️ Tech Stack

- **Lanuage:** [Go 1.25+](https://go.dev/)
- **Web Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **Database ORM:** [Prisma Client Go](https://github.com/steebchen/prisma-client-go)
- **Realtime:** [Socket.io](https://github.com/googollee/go-socket.io)
- **Caching/Queue:** [Redis](https://redis.io/)
- **Profile Acceleration:** Redis-backed JSON caching for authenticated account/profile reads
- **Authentication:** JWT (HMAC SHA256) + Google OAuth

## 📂 Project Structure

```text
core/
├── cmd/api/            # Application entry point
├── internal/
│   ├── modules/        # Domain-driven modules (Auth, Cars, Chat, etc.)
│   │   └── <module>/
│   │       ├── dto/    # Data Transfer Objects / Request-Response models
│   │       ├── handler/# HTTP handlers (Gin)
│   │       ├── repo/   # Data access layer (Prisma/Adapters)
│   │       └── service/# Business logic
│   └── platform/       # Cross-cutting concerns (Auth, Redis, Mailer, etc.)
├── docs/               # Documentation & Migration tracking
├── prisma/             # Schema & Database migrations
└── tests/              # Integration and unit tests
```

## 📦 Core Modules

- **🔐 Auth:** Robust JWT-based authentication, Google OAuth integration, and secure password hashing.
- **🚗 Cars:** Comprehensive vehicle management and classification.
- **💬 Chat:** Real-time messaging powered by Socket.io and Redis.
- **👤 Users:** Profile management and owner-specific flows.
- **🧾 Account/Profile:** Unified profile overview, settings, safety, privacy, documents, device trust, and calendar preferences.
- **🚕 Providers:** Driver and bodyguard management system.
- **💳 Payments:** Integration for managing user payment methods.
- **📅 Rentals:** core rental logic and scheduling.
- **🔔 Notifications:** Multi-channel notification delivery (In-app, SMTP).

---

## 🏁 Getting Started

### 1. Prerequisites

- Go 1.25 or higher
- Redis (running locally or via Docker)
- PostgreSQL (database shared with NestJS app)

### 2. Setup

```bash
# Clone the repository and navigate to core
git clone <repo-url>
cd core

# Install dependencies
go mod tidy

# Setup environment variables
cp .env.template .env
# Edit .env with your local configuration
```

### 3. Database Initialization

```bash
# Generate Prisma Client
go run github.com/steebchen/prisma-client-go generate
```

### 4. Running the Application

```bash
# Development mode
go run ./cmd/api

# Build for production
go build -o pistoni-core ./cmd/api
```

The server will start on the port defined in `.env` (default: `4000`).

---

## 📝 Documentation

- **[Endpoint Parity Matrix](docs/endpoint-parity-matrix.md):** Track migration progress of individual REST endpoints.
- **[Chat Architecture](docs/chat-socketio-realtime.md):** Deep-dive into Socket.io and Redis-backed realtime messaging.
- **[VIP Control Plane](docs/chat-vip-control-plane-architecture.md):** Documentation for the high-priority control plane logic.
- **Module Notes:** Detailed implementation details located in `docs/modules-*.md`.

---

## 🤝 Development Guidelines

- **Clean Architecture:** Keep business logic in `service`, HTTP concerns in `handler`, and DB logic in `repo`.
- **Error Handling:** Use the shared `platform/apperr` for consistent API error responses.
- **Validation:** Always validate DTOs using Gin's built-in binding and `validator/v10`.
- **Security:** Sensitive configuration must only be accessed via environment variables.

---

<p align="center">
  Built with ❤️ by the Pistoni Team.
</p>
