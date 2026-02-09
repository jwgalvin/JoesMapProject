# AI Agent Guidelines for GeoPulse

This document provides context and guidelines for AI coding assistants (GitHub Copilot, Cursor, etc.) working on the GeoPulse project.

## Project Overview

**GeoPulse** is a production-ready Go REST API that ingests and serves geospatial event data from the USGS Earthquake GeoJSON feed. Built with Domain-Driven Design (DDD) principles and Clean Architecture.

**Tech Stack:**
- Go 1.25
- SQLite3 with migrations (golang-migrate)
- Gorilla Mux (HTTP routing)
- testify (testing framework)
- golangci-lint (linting)

**Key Goals:**
- Production-ready code quality (>80% test coverage)
- Clean Architecture with clear layer boundaries
- RESTful API best practices
- Interview-ready portfolio project

## Architecture

### Layer Structure (Clean Architecture + DDD)

```
internal/
├── domain/          # Enterprise business rules (entities, value objects, repository interfaces)
├── application/     # Use cases / application business rules
├── infrastructure/  # External dependencies (database, USGS API, HTTP adapters)
└── interfaces/      # Controllers / API handlers
```

**Dependency Rule:** Dependencies point inward only:
- `interfaces` → `application` → `domain`
- `infrastructure` implements interfaces defined in `domain`

### Domain Layer (`internal/domain/`)
- **Entities:** Core business objects with identity (e.g., `Event`)
- **Value Objects:** Immutable objects without identity (e.g., `Magnitude`, `Location`, `EventType`)
- **Repository Interfaces:** Define data access contracts (implemented in `infrastructure`)
- **No external dependencies** - pure Go, no frameworks

### Application Layer (`internal/application/`)
- **Use Cases:** Business workflows (e.g., `IngestService`, `QueryService`)
- References domain entities and repository interfaces
- Orchestrates domain logic

### Infrastructure Layer (`internal/infrastructure/`)
- **Implementations:** Concrete implementations of repository interfaces
- **External APIs:** USGS client, HTTP adapters
- **Database:** SQLite persistence with go-sqlite3

### Interfaces Layer (`internal/interfaces/`)
- **HTTP Handlers:** REST API endpoints using Gorilla Mux
- Maps HTTP requests to application use cases
- Returns JSON responses

## Code Conventions

### General Go Guidelines
- Follow **Effective Go** and **Go Code Review Comments**
- Use `gofmt` formatting (run `make fmt`)
- Pass `go vet` (run `make vet`)
- Pass golangci-lint with all configured linters (run `make lint`)
- All exported functions/types must have godoc comments

### Naming Conventions
- **Packages:** Short, lowercase, singular nouns (e.g., `event`, `query`, `http`)
- **Files:** Lowercase with underscores (e.g., `event_repository.go`, `magnitude_test.go`)
- **Interfaces:** Suffix with "-er" when possible (e.g., `EventRepository`, `Ingester`)
- **Tests:** `_test.go` suffix, test functions start with `Test`

### Error Handling
- Return errors, don't panic (except in truly exceptional cases)
- Wrap errors with context: `fmt.Errorf("failed to save event: %w", err)`
- Domain errors should be defined in domain layer as custom types
- Use `errors.Is()` and `errors.As()` for error checking

### Testing
- **Unit tests:** Test files alongside code (`*_test.go`)
- **Integration tests:** In `tests/integration/`
- Use testify for assertions: `assert`, `require`
- Aim for >70% coverage
- Table-driven tests for multiple scenarios
- Test file structure:
  ```go
  func TestFunctionName(t *testing.T) {
      // Arrange
      // Act
      // Assert
  }
  ```

### Value Objects
- Always immutable
- Validate in constructors (factory functions like `NewMagnitude()`)
- Return errors for invalid values
- Implement `String()` method for debugging
- Example:
  ```go
  type Magnitude struct {
      value float64
      scale string
  }
  
  func NewMagnitude(value float64, scale string) (Magnitude, error) {
      if value < 0 {
          return Magnitude{}, fmt.Errorf("magnitude cannot be negative: %f", value)
      }
      return Magnitude{value: value, scale: scale}, nil
  }
  ```

### Repository Pattern
- Define interface in `domain/` package
- Implement in `infrastructure/persistence/` package
- Use dependency injection (pass repository to use cases)
- Example from domain:
  ```go
  type EventRepository interface {
      Save(ctx context.Context, event Event) error
      FindByID(ctx context.Context, id string) (Event, error)
      FindAll(ctx context.Context, filters Filters) ([]Event, error)
  }
  ```

## Database

### Migrations
- Use golang-migrate with SQLite
- Files in `migrations/` directory
- Naming: `000001_create_events_table.up.sql` and `.down.sql`
- Run with: `make migrate-up`, `make migrate-down`

### Schema Conventions
- Table names: plural, lowercase with underscores (e.g., `events`)
- Primary keys: `id` (TEXT for UUIDs or external IDs)
- Timestamps: `created_at`, `updated_at` (DATETIME)
- Use foreign keys with `ON DELETE CASCADE` when appropriate

## API Design

### REST Conventions
- Use standard HTTP methods: GET, POST, PUT, PATCH, DELETE
- Resource-based URLs: `/api/v1/events/{id}`
- Return appropriate status codes (200, 201, 400, 404, 500)
- JSON request/response bodies
- Use pagination for list endpoints

### Response Format
```json
{
  "data": { ... },
  "error": null,
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100
  }
}
```

### Error Response Format
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid magnitude value",
    "details": { ... }
  }
}
```

## Configuration

- **configs/.env:** Secrets and environment-specific vars (DATABASE_URL, PORT, API keys)
- **configs/config.yaml:** Application config (rate limits, timeouts, external endpoints)
- Load with `godotenv` and `yaml.v3`
- Never commit `.env` (use `.env.example` for templates)

## Development Workflow

### Before Committing
```powershell
make fmt      # Format code
make vet      # Run go vet
make lint     # Run golangci-lint
make test     # Run tests
```

Or run all at once:
```powershell
make check    # Runs fmt + vet + lint
make ci       # Runs check + test
```

### Adding a New Feature
1. Start in **domain layer** - define entities, value objects, repository interfaces
2. Write **tests first** for domain logic (TDD encouraged)
3. Implement **use case** in application layer
4. Create **repository implementation** in infrastructure layer
5. Add **HTTP handler** in interfaces layer
6. Write **integration tests** in `tests/integration/`
7. Update **API documentation** in `docs/api.md`

### Adding a New Endpoint
1. Define handler in `internal/interfaces/api/`
2. Register route in router setup (likely `cmd/api/main.go` or dedicated router file)
3. Implement handler logic (validate request, call use case, return response)
4. Add tests for handler
5. Document in `docs/api.md`

## Common Patterns

### Dependency Injection
```go
// In main.go or similar
db := setupDatabase()
eventRepo := persistence.NewSQLiteEventRepository(db)
ingestService := ingest.NewService(eventRepo, usgsClient)
handler := api.NewIngestHandler(ingestService)
```

### Context Usage
- Always pass `context.Context` as first parameter
- Use for cancellation, timeouts, request-scoped values
- Example: `func (r *Repository) Save(ctx context.Context, event Event) error`

### Logging
- Use structured logging (consider adding `slog` or `zerolog` in future)
- Log at appropriate levels: Debug, Info, Warn, Error
- Include relevant context in log messages

## File Organization

### Typical Package Structure
```
internal/domain/event/
├── event.go           # Event entity
├── event_test.go      # Event tests
├── magnitude.go       # Magnitude value object
├── magnitude_test.go  # Magnitude tests
├── location.go        # Location value object
├── location_test.go   # Location tests
├── repository.go      # EventRepository interface
└── event_type.go      # EventType enum/value object
```

## AI Assistant Guidelines

### When Writing Code
1. **Follow the architecture** - respect layer boundaries
2. **Write tests** - include test files with implementation
3. **Validate inputs** - especially in value object constructors
4. **Use context** - always include `context.Context` in function signatures for I/O
5. **Document exports** - add godoc comments for all exported symbols
6. **Handle errors properly** - wrap with context, return to caller

### When Refactoring
1. **Run tests first** - ensure they pass before changes
2. **Make small changes** - incremental refactoring is safer
3. **Update tests** - keep tests in sync with code changes
4. **Check linters** - run `make lint` after refactoring

### When Adding Dependencies
1. **Justify the dependency** - prefer standard library when possible
2. **Update go.mod** - use `go get <package>`
3. **Run `make deps`** - ensure dependencies download
4. **Update AGENTS.md** - document major dependencies in Tech Stack section

### Code Review Checklist
- [ ] Follows Clean Architecture layer boundaries
- [ ] All exports have godoc comments
- [ ] Tests included and passing (`make test`)
- [ ] No linter errors (`make lint`)
- [ ] Error handling with proper wrapping
- [ ] Context passed to I/O operations
- [ ] No hardcoded values (use config)
- [ ] Immutable value objects with validation

## Resources

- [Step-by-step Implementation Guide](IMPLEMENTATION_GUIDE.md)
- [Project Overview](ProjectOverview.md)
- [API Documentation](docs/api.md)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Clean Architecture (Uncle Bob)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## Current Status

**Completed:**
- ✅ Project structure and scaffolding (Step 01)
- ✅ Build tooling (Makefile, linting, testing)
- ✅ Configuration management (.env, config.yaml)

**Next Steps:**
- ⏳ Domain model implementation (Step 02)
- ⏳ Repository interfaces and implementations
- ⏳ Use case/application services
- ⏳ HTTP handlers and routing
- ⏳ Database migrations
- ⏳ Integration tests
- ⏳ Docker and deployment

See [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) for detailed step-by-step plan.

---

**Remember:** This is a portfolio project that should demonstrate production-ready Go code and architectural best practices. Code quality matters more than speed!
