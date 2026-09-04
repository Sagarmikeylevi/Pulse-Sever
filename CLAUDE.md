# Pulse Server

Productivity tracking backend API — Go, Gin, GORM, PostgreSQL.

## Commands

```bash
make run                          # start server (auto-runs migrations)
make test                         # run all tests
make db-up                        # start PostgreSQL
make migrate-create name=xyz      # create new migration files
```

## Architecture

```
Routes → Controller → Service → Repository → Entity
```

Each layer only talks to the layer below it. Repositories and services use interfaces for testability.

## Conventions

- UUID primary keys, soft deletes on core tables
- Migrations: raw SQL via golang-migrate, NOT GORM AutoMigrate
- DTOs in `internal/dto/`, mocks in `internal/mocks/`
- Test files use `package <name>_test` to avoid import cycles
- API endpoints under `/api/v1/`

## Feature Docs

Read `docs/` before modifying a feature — they capture design decisions and flows.

- [docs/auth.md](docs/auth.md) — Authentication system
