# Step 11: Docker and Deployment

## Objective
Create Docker configuration and deploy to a cloud platform (Fly.io, Render, or Railway).

## Tasks

### 1. Create Dockerfile
Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o geopulse ./cmd/api

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/geopulse .

# Copy migrations
COPY migrations ./migrations

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV DATABASE_PATH=/data/geopulse.db
ENV LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/v1/health || exit 1

# Run the binary
CMD ["./geopulse"]
```

### 2. Create .dockerignore
Create `.dockerignore`:
```
# Git
.git
.gitignore

# Development
.env
.env.local
*.db
*.db-shm
*.db-wal
data/

# Testing
coverage.out
coverage.html
*_test.go
tests/

# Documentation
*.md
docs/

# Build artifacts
geopulse
geopulse.exe
bin/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
```

### 3. Create Docker Compose for Local Development
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - PORT=8080
      - DATABASE_PATH=/data/geopulse.db
      - LOG_LEVEL=info
      - USGS_POLL_INTERVAL=5
      - ENABLE_CORS=true
      - ALLOWED_ORIGINS=http://localhost:3000
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/v1/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s
```

### 4. Test Docker Build
```powershell
# Build image
docker build -t geopulse:latest .

# Run container
docker run -p 8080:8080 -v ${PWD}/data:/data geopulse:latest

# Test with docker-compose
docker-compose up

# Stop containers
docker-compose down
```

### 5. Deploy to Fly.io

#### Install Fly CLI
```powershell
# Windows
iwr https://fly.io/install.ps1 -useb | iex

# Verify installation
fly version
```

#### Login and Initialize
```powershell
# Login to Fly.io
fly auth login

# Launch app (interactive setup)
fly launch

# When prompted:
# - Choose app name (e.g., geopulse-api)
# - Choose region
# - Don't deploy yet
```

#### Configure Fly.io
Create `fly.toml`:
```toml
app = "geopulse-api"
primary_region = "sjc"

[build]
  [build.args]
    GO_VERSION = "1.22"

[env]
  PORT = "8080"
  LOG_LEVEL = "info"
  USGS_POLL_INTERVAL = "5"
  ENABLE_CORS = "true"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0

  [[http_service.checks]]
    interval = "30s"
    timeout = "5s"
    grace_period = "10s"
    method = "GET"
    path = "/v1/health"

[[vm]]
  cpu_kind = "shared"
  cpus = 1
  memory_mb = 256

[[mounts]]
  source = "geopulse_data"
  destination = "/data"
  initial_size = "1gb"
```

#### Create Volume and Deploy
```powershell
# Create persistent volume for SQLite
fly volumes create geopulse_data --size 1

# Deploy
fly deploy

# Check status
fly status

# View logs
fly logs

# Open in browser
fly open /v1/health
```

### 6. Alternative: Deploy to Render

Create `render.yaml`:
```yaml
services:
  - type: web
    name: geopulse-api
    env: docker
    region: oregon
    plan: free
    healthCheckPath: /v1/health
    envVars:
      - key: PORT
        value: 8080
      - key: DATABASE_PATH
        value: /data/geopulse.db
      - key: LOG_LEVEL
        value: info
      - key: USGS_POLL_INTERVAL
        value: 5
    disk:
      name: geopulse-data
      mountPath: /data
      sizeGB: 1
```

Deploy:
1. Push to GitHub
2. Go to [render.com](https://render.com)
3. Connect GitHub repository
4. Render will auto-deploy using `render.yaml`

### 7. Alternative: Deploy to Railway

Create `railway.json`:
```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "./Dockerfile"
  },
  "deploy": {
    "numReplicas": 1,
    "healthcheckPath": "/v1/health",
    "restartPolicyType": "ON_FAILURE"
  }
}
```

Deploy:
```powershell
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Initialize project
railway init

# Link to project
railway link

# Deploy
railway up

# View logs
railway logs
```

### 8. Set Environment Variables (Production)

For Fly.io:
```powershell
fly secrets set USGS_ENDPOINT="https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson"
fly secrets set ALLOWED_ORIGINS="https://yourdomain.com"
```

For Render/Railway: Set in dashboard UI.

### 9. Create Deployment Documentation
Create `docs/deployment.md`:
```markdown
# Deployment Guide

## Prerequisites
- Docker installed
- Account on deployment platform (Fly.io/Render/Railway)

## Local Docker Deployment

```bash
docker-compose up -d
```

Access API at http://localhost:8080

## Production Deployment

### Fly.io

1. Install Fly CLI
2. Login: `fly auth login`
3. Deploy: `fly deploy`
4. Check status: `fly status`

### Render

1. Push to GitHub
2. Connect repository on Render dashboard
3. Render auto-deploys from `render.yaml`

### Railway

1. Install Railway CLI: `npm i -g @railway/cli`
2. Login: `railway login`
3. Deploy: `railway up`

## Post-Deployment

### Verify Deployment
```bash
curl https://your-app.fly.dev/v1/health
```

### View Logs
```bash
# Fly.io
fly logs

# Render
Via dashboard

# Railway
railway logs
```

### Scale
```bash
# Fly.io
fly scale count 2
fly scale memory 512

# Render/Railway
Via dashboard
```

## Database Backups

SQLite database is persisted on volume. Backup recommendations:
- Schedule periodic volume snapshots
- Export data periodically
- Use platform backup features

## Monitoring

- Health check: `/v1/health`
- View logs for ingestion metrics
- Monitor response times
- Track error rates
```

### 10. Test Production Deployment

```powershell
# Test health endpoint
curl https://your-app-url/v1/health

# Test events endpoint
curl https://your-app-url/v1/events?limit=5

# Test GeoJSON
curl https://your-app-url/v1/events/geojson?minMagnitude=4.0
```

## Success Criteria
- ✓ Dockerfile created and tested
- ✓ Docker Compose working locally
- ✓ Application deployed to cloud platform
- ✓ Persistent storage configured
- ✓ Health checks passing
- ✓ Environment variables configured
- ✓ HTTPS enabled
- ✓ API accessible publicly

## Next Step
Proceed to **Step12-Documentation.md** to finalize project documentation.
