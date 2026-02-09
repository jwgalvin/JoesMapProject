# Step 10: CI/CD Setup with GitHub Actions

## Objective
Configure GitHub Actions for continuous integration and optionally continuous deployment.

## Tasks

### 1. Create GitHub Actions Workflow Directory
```powershell
New-Item -ItemType Directory -Path .github\workflows -Force
```

### 2. Create CI Workflow
Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test ./... -v -coverprofile=coverage.out -covermode=atomic

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          fail_ci_if_error: false

  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build binary
        run: go build -v -o geopulse ./cmd/api

      - name: Upload binary
        uses: actions/upload-artifact@v4
        with:
          name: geopulse-binary
          path: geopulse
```

### 3. Create golangci-lint Configuration
Create `.golangci.yml`:
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - exportloopref
    - gocyclo
    - gocritic
    - misspell

linters-settings:
  gocyclo:
    min-complexity: 15
  
  gocritic:
    enabled-checks:
      - appendAssign
      - assignOp
      - boolExprSimplify
      - dupArg
      - dupBranchBody
      - dupCase
      - dupSubExpr
      - elseif
      - emptyFallthrough
      - emptyStringTest
      - equalFold
      - flagDeref
      - ifElseChain
      - methodExprCall
      - nilValReturn
      - octalLiteral
      - rangeExprCopy
      - rangeValCopy
      - regexpMust
      - stringXbytes
      - typeAssertChain
      - typeSwitchVar
      - underef
      - unlambda
      - unslice

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
```

### 4. Create Docker Build Workflow (Optional)
Create `.github/workflows/docker.yml`:
```yaml
name: Docker Build

on:
  push:
    branches: [ main ]
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Log in to the Container registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
```

### 5. Add Status Badges to README
Update `README.md` to include status badges:
```markdown
# GeoPulse - Go Geospatial Events API

[![CI](https://github.com/yourusername/geopulse/workflows/CI/badge.svg)](https://github.com/yourusername/geopulse/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/geopulse)](https://goreportcard.com/report/github.com/yourusername/geopulse)
[![codecov](https://codecov.io/gh/yourusername/geopulse/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/geopulse)

Real-time geospatial event tracking API built with Go.

[Rest of README content...]
```

### 6. Initial Git Commit
```powershell
# Initialize and commit
git add .
git commit -m "Initial commit: GeoPulse API implementation"

# Create main branch (if needed)
git branch -M main

# Add remote (replace with your GitHub repo URL)
git remote add origin https://github.com/yourusername/geopulse.git

# Push to GitHub
git push -u origin main
```

### 7. Verify CI Pipeline
After pushing to GitHub:

1. Go to your repository on GitHub
2. Click on **Actions** tab
3. Verify that the CI workflow runs successfully
4. Check test results and coverage

### 8. Set Up Branch Protection (Recommended)
On GitHub:

1. Go to **Settings** → **Branches**
2. Add rule for `main` branch:
   - Require pull request reviews
   - Require status checks to pass (CI workflow)
   - Require branches to be up to date

### 9. Create Pull Request Template (Optional)
Create `.github/pull_request_template.md`:
```markdown
## Description
<!-- Describe your changes in detail -->

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
<!-- Describe the tests you ran -->
- [ ] All tests pass locally
- [ ] Added new tests for new functionality
- [ ] Updated documentation

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
```

### 10. Local Linting
Run linter locally before committing:

```powershell
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

## Success Criteria
- ✓ GitHub Actions workflows created
- ✓ CI pipeline runs on push/PR
- ✓ Tests run automatically
- ✓ Linting configured
- ✓ Coverage reporting set up
- ✓ Build artifacts generated
- ✓ Branch protection enabled

## Next Step
Proceed to **Step11-DockerDeployment.md** to containerize and deploy the application.
