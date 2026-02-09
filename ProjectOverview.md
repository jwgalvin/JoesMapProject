# GeoPulse – Go Geospatial Events API

## Project Overview

**GeoPulse** is a Go-based backend service that ingests real-world geospatial event data from free external APIs, normalizes it using Domain-Driven Design (DDD), and exposes a clean, queryable REST API suitable for mapping and alerting use cases.

The initial milestone focuses on:

* API-only (no frontend yet)
* Test-Driven Development (TDD)
* Clean domain modeling
* CI/CD via GitHub Actions
* Low-cost cloud deployment

The API is designed to be **map-ready**, returning GeoJSON-compatible data so a mapping UI (e.g., Leaflet) can be added later without refactoring.

---

## Goals

* Practice Go in an interview-realistic context
* Demonstrate backend fundamentals (API design, testing, data modeling)
* Follow Domain-Driven Design principles
* Integrate with a real external API
* Use GitHub Actions for CI
* Deploy to a free or low-cost cloud platform

---

## External Data Sources

### Primary API: USGS Earthquake Feed

* Free, public, no authentication required
* GeoJSON format
* Real-time geospatial data

Example endpoint:

```
https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson
```

The service will periodically fetch this data and normalize it into internal domain entities.

---

## Configuration Management

### Environment Variables

* `PORT` - HTTP server port (default: 8080)
* `DATABASE_PATH` - SQLite database file location
* `USGS_POLL_INTERVAL` - Polling frequency in minutes (default: 5)
* `LOG_LEVEL` - Logging level (debug, info, warn, error)
* `ENABLE_CORS` - Enable CORS for frontend integration

### Local Development

* Use `godotenv` to load `.env` file
* Provide `.env.example` for team setup
* Sensitive configs never committed to Git

### Configuration File

Optional `config.yaml` for:

* API rate limits
* Query parameter bounds (max radius, max results)
* External API endpoints and timeouts

---

## Domain Model (DDD)

### Core Domain: Geospatial Events

An **Event** represents something that occurred at a specific place and time.

#### Event Properties

* ID
* EventType (earthquake, storm, etc.)
* Source (e.g., "USGS")
* OccurredAt (timestamp)
* Location (latitude, longitude)
* Magnitude
* Depth (for earthquakes, in km)
* Metadata (raw source ID, optional fields)
* CreatedAt (audit timestamp)
* UpdatedAt (audit timestamp)

### Value Objects

* `Location` (lat, lng with validation)
* `Magnitude` (with min/max bounds)
* `EventID` (unique identifier)
* `EventType` (enumeration of supported event types)

### Repository Interface

The domain defines repository interfaces without knowing storage details.

---

## Project Structure

```
/cmd
  /api
    main.go

/internal
  /domain
    /event
      event.go
      location.go
      magnitude.go
      event_type.go
      repository.go

  /application
    /ingest
      ingest_service.go
      scheduler.go
    /query
      query_service.go

  /infrastructure
    /usgs
      client.go
    /persistence
      sqlite_repository.go
    /http
      handlers.go
      router.go
      middleware.go

  /interfaces
    /api
      dto.go

/configs
  config.yaml
  .env.example

/migrations
  001_create_events_table.sql
  002_add_indexes.sql

/scripts
  build.sh
  deploy.sh

/docs
  api.md

/tests
  /integration
  /unit
```

---

## Application Flow

### Ingest Flow

1. Background scheduler triggers every 5 minutes (configurable)
2. USGS client fetches GeoJSON data
3. Data is decoded and validated
4. External records are mapped to domain `Event`
5. Duplicate detection using source event IDs
6. Events are upserted via repository interface (insert new, update existing)
7. Structured logging of ingestion metrics

### Query Flow

1. HTTP handler parses query parameters
2. Application query service applies filters
3. Results are mapped to API DTOs
4. Response is returned as JSON or GeoJSON

---

## Database

### Initial Choice: SQLite

* Simple, file-based
* No infrastructure cost
* Perfect for local dev and small deployments

Table: `events`

Fields:

* id (PRIMARY KEY)
* event_type (VARCHAR)
* source (VARCHAR)
* source_event_id (VARCHAR, UNIQUE for duplicate detection)
* occurred_at (TIMESTAMP)
* latitude (REAL)
* longitude (REAL)
* magnitude (REAL)
* depth (REAL, nullable)
* metadata (JSON/TEXT)
* created_at (TIMESTAMP)
* updated_at (TIMESTAMP)

### Indexes

* Index on `occurred_at` for time-range queries
* Index on `magnitude` for filtering
* Composite index on `(latitude, longitude)` for spatial queries
* Unique index on `(source, source_event_id)`

### Migrations

Using `golang-migrate` for versioned schema changes:

* `001_create_events_table.sql`
* `002_add_indexes.sql`

The persistence layer is isolated so SQLite can later be replaced with Postgres without touching domain logic.

---

## API Design

### API Versioning

All endpoints prefixed with `/v1` for future compatibility:

```
GET /v1/health
GET /v1/events
GET /v1/events/geojson
```

### Core Endpoints

#### Health Check

```
GET /v1/health
```

Returns service status and basic metrics for deployment platforms.

#### List Events

```
GET /v1/events
```

### Supported Query Parameters

* `minMagnitude` (float, min: 0, max: 10)
* `maxMagnitude` (float, min: 0, max: 10)
* `lat` (float, -90 to 90)
* `lng` (float, -180 to 180)
* `radiusKm` (float, min: 1, max: 20000)
* `fromTime` (ISO 8601 timestamp)
* `toTime` (ISO 8601 timestamp)
* `eventType` (string: earthquake, storm, etc.)
* `limit` (int, default: 100, max: 1000)
* `offset` (int, default: 0)

### Response Format

#### Standard JSON Response

```json
{
  "data": [...],
  "pagination": {
    "limit": 100,
    "offset": 0,
    "total": 250
  }
}
```

#### Error Response

```json
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "minMagnitude must be between 0 and 10",
    "field": "minMagnitude"
  }
}
```

### GeoJSON Output

The `/v1/events/geojson` endpoint returns:

* FeatureCollection
* Point geometries
* Event properties (magnitude, timestamp, source, depth)
* Accepts same query parameters as `/v1/events`

This enables immediate compatibility with mapping libraries later.

### CORS Configuration

Configurable CORS middleware:

* Allow specific origins for frontend
* Support preflight requests
* Configurable via environment variable

---

## Testing Strategy (TDD)

### Domain Tests

* Event validation
* Location bounds
* Magnitude rules

### Application Tests

* Ingest service with mocked USGS client
* Query filtering logic

### Infrastructure Tests

* USGS client decoding
* SQLite repository behavior
* HTTP handlers using `httptest`

External APIs are always mocked in tests.

### Integration Tests

* Full request-to-database flow
* API endpoint validation with real SQLite database
* Ingest service with mocked external API
* Query filtering with actual persistence

### Performance Considerations

* Load testing for query endpoints (optional)
* Verify pagination handles large datasets
* Ensure indexes improve query performance

---

## Operational Concerns

### Logging

* Use structured logging (`log/slog` or `zap`)
* Log levels: DEBUG, INFO, WARN, ERROR
* Include request IDs for tracing
* Log ingestion metrics (events processed, failures)

### Observability

* Event count metrics
* API request metrics (response times, status codes)
* Database query performance
* Ingestion job success/failure rates

### Graceful Shutdown

* Handle SIGTERM/SIGINT signals
* Complete in-flight HTTP requests
* Close database connections cleanly
* Stop background ingestion jobs

### Background Job Scheduling

* Use `time.Ticker` for periodic USGS polling
* Configurable poll interval (default: 5 minutes)
* Prevent concurrent job execution
* Structured logging of job execution

---

## Security Considerations

### Input Validation

* Validate all query parameters with bounds checking
* Sanitize user inputs
* Return 400 Bad Request for invalid inputs

### Rate Limiting

* Implement per-IP rate limiting middleware
* Configurable limits (e.g., 100 req/min)
* Return 429 Too Many Requests when exceeded

### SQL Injection Prevention

* Use parameterized queries exclusively
* Repository layer handles all SQL construction
* Never interpolate user input into SQL strings

### HTTPS

* Deployment platforms handle TLS termination
* Redirect HTTP to HTTPS in production

---

## CI/CD (GitHub Actions)

### Continuous Integration

Pipeline runs on every pull request:

* `go test ./...`
* `go vet ./...`
* `golangci-lint run`

### Continuous Deployment (Optional)

* Build Docker image on main branch
* Push to GitHub Container Registry
* Auto-deploy to Fly.io on success

---

## Docker

### Dockerfile

Multi-stage build:

1. Build stage: Compile Go binary
2. Runtime stage: Alpine Linux with binary only

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api ./cmd/api

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/api /api
EXPOSE 8080
CMD ["/api"]
```

### Docker Compose (Local Dev)

Optional for local development convenience:

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    env_file:
      - .env
```

---

## Deployment

### Target Platforms (Free / Low Cost)

* Fly.io (preferred)
* Render
* Railway

Deployment runs:

* Go binary or Docker container
* SQLite database persisted via volume
* Environment variables configured in platform dashboard
* Health check endpoint monitors service status

### Platform-Specific Notes

**Fly.io:**
* Use `fly.toml` configuration
* Persistent volumes for SQLite
* Automatic HTTPS

**Render:**
* Native Go support
* Disk persistence available
* Free tier limitations on database size

---

## Documentation

### API Documentation

* Maintain `docs/api.md` with endpoint specifications
* Include example requests and responses
* Document error codes and meanings
* Optional: Generate OpenAPI/Swagger spec later

### Development Setup

README includes:

* Go version requirement (1.22+)
* Clone and setup instructions
* Environment variable configuration
* Running tests
* Local development workflow

---

## Milestones

### Milestone 1 – Core API

* Domain model
* USGS ingestion
* SQLite persistence
* Queryable REST API
* CI pipeline
* Cloud deployment

### Milestone 2 – Mapping UI

* Static frontend (Leaflet)
* Consume `/events/geojson`
* Render markers and popups

### Milestone 3 – Extensions

* Additional data sources (NOAA, NASA)
* Alerting thresholds
* Background ingestion jobs

---

## How to Run (Planned)

```
go run ./cmd/api
```

```
curl http://localhost:8080/events
```

---

## Notes

This project intentionally favors clarity and correctness over complexity. Each layer is designed to evolve independently as requirements grow.
