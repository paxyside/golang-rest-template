# My Go Microservice Template

[![Use this template](https://img.shields.io/badge/-Use%20this%20template-brightgreen?style=for-the-badge)](https://github.com/paxyside/golang-rest-template/generate)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/paxyside/golang-rest-template)

A basic Go microservice template designed for rapid development:
- REST API with Gin
- PostgreSQL integration
- Database migrations
- Prometheus metrics
- RabbitMQ message broker
- Swagger documentation
- Built-in linting, dockerization, and graceful shutdown

> This template is intended for starting new Go projects with a ready-to-use infrastructure setup.

## 🚀 Features

- ⚡ Clean Architecture project structure
- 🛡️ HTTP server powered by [Gin Web Framework](https://github.com/gin-gonic/gin)
- 📄 Auto-generated Swagger API documentation
- 📂 PostgreSQL support with [PGX Pool](https://github.com/jackc/pgx)
- 🐇 Asynchronous messaging with [RabbitMQ](https://github.com/rabbitmq/amqp091-go)
- 📈 Monitoring with Prometheus metrics
- 🐳 Ready-to-use Docker and docker-compose setup
- 🔧 Makefile with handy development commands
- 🧹 GolangCI-Lint integrated for code quality control

## 📚 Installed Packages

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Gin-Prometheus Middleware](https://github.com/zsais/go-gin-prometheus)
- [PGX (PostgreSQL driver)](https://github.com/jackc/pgx)
- [Go-Swagger3](https://github.com/parvez3019/go-swagger3)
- [Go-Migrate](https://github.com/golang-migrate/migrate)
- [RabbitMQ Client (AMQP 0-9-1)](https://github.com/rabbitmq/amqp091-go)

## 🛠️ Quickstart

### 1. Clone the repository
```bash
git clone https://github.com/paxyside/project_reference.git my_project
cd my_project
```

### 2. Set up environment variables
```bash
cp .env.example .env
chmod 600 .env
```

### 3. Create Docker network (first time only)
```bash
docker network create {network_name}
```

## 📋 Notes

- Update your `.env` file with the correct database and RabbitMQ credentials.
- The template uses graceful shutdown handlers for clean exits.
- Extend the `infrastructure/` and `internal/` packages for additional services (e.g., Redis, Kafka).

## 🧹 TODO (optional improvements)

- Add Redis infrastructure
- Add Healthcheck endpoint
- Improve Swagger schema auto-generation
- Add tracing (Jaeger/OpenTelemetry)

## 📈 Monitoring and Tools

- **Swagger UI** — REST API Documentation
- **Prometheus UI** — Metrics Monitoring
- **RabbitMQ UI** — RabbitMQ Management Console

## 🗂️ Project Structure

```
├── cmd/
│   └── server/                  # Application entry point (main.go)
├── docs/
│   └── swagger.json              # Swagger API documentation
├── http/
│   └── requests.http             # HTTP request examples for testing
├── infrastructure/
│   ├── config/                   # Application configuration parsing
│   ├── database/                 # Database connection and initialization
│   └── rabbit/                   # RabbitMQ client setup
├── internal/
│   ├── application/              # Core application logic (server setup, shutdown handling)
│   ├── controller/               # HTTP controllers, routes, and middleware
│   ├── domain/                   # Domain models and interfaces
│   ├── persistence/              # Repositories and data persistence logic
│   ├── service/                  # Business services
│   └── worker/                   # Asynchronous background workers
├── migrations/                   # SQL database migration scripts
├── docker-compose.yaml           # Docker Compose configuration
├── Dockerfile                    # Application Dockerfile
├── Makefile                      # Common build/run commands
├── prometheus.yaml               # Prometheus monitoring configuration
└── README.md                     # Project documentation
```

