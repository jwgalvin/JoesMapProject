#!/usr/bin/env pwsh
# PowerShell build script for GeoPulse (Windows-native alternative to Makefile)
# Usage: .\make.ps1 <command>
# Example: .\make.ps1 test
# Example: .\make.ps1 build

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Variables
$BinaryName = "geopulse"
$MainPath = "./cmd/api"
$CoverageFile = "coverage.out"
$CoverageHtml = "coverage.html"

function Show-Help {
    Write-Host "GeoPulse Build Commands:" -ForegroundColor Cyan
    Write-Host "  .\make.ps1 build          - Build the application" -ForegroundColor White
    Write-Host "  .\make.ps1 run            - Run the application" -ForegroundColor White
    Write-Host "  .\make.ps1 test           - Run all tests" -ForegroundColor White
    Write-Host "  .\make.ps1 test-coverage  - Run tests with coverage report" -ForegroundColor White
    Write-Host "  .\make.ps1 lint           - Run golangci-lint" -ForegroundColor White
    Write-Host "  .\make.ps1 fmt            - Format code with gofmt" -ForegroundColor White
    Write-Host "  .\make.ps1 vet            - Run go vet" -ForegroundColor White
    Write-Host "  .\make.ps1 clean          - Remove build artifacts" -ForegroundColor White
    Write-Host "  .\make.ps1 deps           - Download dependencies" -ForegroundColor White
    Write-Host "  .\make.ps1 tidy           - Tidy and verify dependencies" -ForegroundColor White
    Write-Host "  .\make.ps1 check          - Run all checks (fmt, vet, lint)" -ForegroundColor White
    Write-Host "  .\make.ps1 ci             - Run full CI pipeline" -ForegroundColor White
}

function Invoke-Build {
    Write-Host "Building $BinaryName..." -ForegroundColor Green
    go build -v -o "$BinaryName.exe" $MainPath
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "Build successful: $BinaryName.exe" -ForegroundColor Green
}

function Invoke-Run {
    Write-Host "Running $BinaryName..." -ForegroundColor Green
    go run $MainPath
}

function Invoke-Test {
    Write-Host "Running tests..." -ForegroundColor Green
    go test ./internal/... -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed!" -ForegroundColor Red
        exit 1
    }
}

function Invoke-TestCoverage {
    Write-Host "Running tests with coverage..." -ForegroundColor Green
    
    # Run tests with coverage
    go test ./internal/... -coverprofile=$CoverageFile -covermode=atomic
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed!" -ForegroundColor Red
        exit 1
    }
    
    # Generate HTML report
    go tool cover -html=$CoverageFile -o $CoverageHtml
    Write-Host "Coverage report generated: $CoverageHtml" -ForegroundColor Green
    
    # Show total coverage
    $coverageOutput = go tool cover -func=$CoverageFile | Select-String "total"
    Write-Host $coverageOutput -ForegroundColor Cyan
    
    # Extract coverage percentage and check threshold
    $coverageLine = go tool cover -func=$CoverageFile | Select-String "total" | Out-String
    if ($coverageLine -match "(\d+\.\d+)%") {
        $coveragePercent = [double]$matches[1]
        $threshold = 70.0
        
        Write-Host "Coverage: $coveragePercent%" -ForegroundColor Cyan
        
        if ($coveragePercent -lt $threshold) {
            Write-Host "Coverage $coveragePercent% is below threshold $threshold%" -ForegroundColor Red
            exit 1
        }
        Write-Host "Coverage meets threshold ($threshold%)" -ForegroundColor Green
    }
}

function Invoke-Lint {
    Write-Host "Running golangci-lint..." -ForegroundColor Green
    
    # Check if golangci-lint is installed
    $lintInstalled = Get-Command golangci-lint -ErrorAction SilentlyContinue
    if (-not $lintInstalled) {
        Write-Host "golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/" -ForegroundColor Yellow
        Write-Host "Or run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" -ForegroundColor Yellow
        exit 1
    }
    
    golangci-lint run --timeout=5m
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Lint failed!" -ForegroundColor Red
        exit 1
    }
}

function Invoke-Format {
    Write-Host "Formatting code..." -ForegroundColor Green
    go fmt ./...
}

function Invoke-Vet {
    Write-Host "Running go vet..." -ForegroundColor Green
    go vet ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Vet failed!" -ForegroundColor Red
        exit 1
    }
}

function Invoke-Clean {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Green
    
    # Remove binaries
    if (Test-Path "$BinaryName.exe") { Remove-Item "$BinaryName.exe" -Force }
    if (Test-Path $BinaryName) { Remove-Item $BinaryName -Force }
    
    # Remove coverage files
    if (Test-Path $CoverageFile) { Remove-Item $CoverageFile -Force }
    if (Test-Path $CoverageHtml) { Remove-Item $CoverageHtml -Force }
    
    # Remove database files
    Get-ChildItem -Filter "*.db" -Recurse | Remove-Item -Force
    Get-ChildItem -Filter "*.db-shm" -Recurse | Remove-Item -Force
    Get-ChildItem -Filter "*.db-wal" -Recurse | Remove-Item -Force
    
    # Remove data directory (except .gitkeep)
    if (Test-Path "data") {
        Get-ChildItem "data" -Exclude ".gitkeep" | Remove-Item -Force -Recurse
    }
    
    Write-Host "Clean complete!" -ForegroundColor Green
}

function Invoke-Deps {
    Write-Host "Downloading dependencies..." -ForegroundColor Green
    go mod download
}

function Invoke-Tidy {
    Write-Host "Tidying dependencies..." -ForegroundColor Green
    go mod tidy
    go mod verify
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Dependency verification failed!" -ForegroundColor Red
        exit 1
    }
}

function Invoke-Check {
    Write-Host "Running all checks..." -ForegroundColor Green
    Invoke-Format
    Invoke-Vet
    Invoke-Lint
    Invoke-Test
    Write-Host "All checks passed!" -ForegroundColor Green
}

function Invoke-CI {
    Write-Host "Running CI pipeline..." -ForegroundColor Green
    Invoke-Deps
    Invoke-Tidy
    Invoke-Check
    Invoke-TestCoverage
    Write-Host "CI pipeline completed!" -ForegroundColor Green
}

# Main command dispatcher
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "build" { Invoke-Build }
    "run" { Invoke-Run }
    "test" { Invoke-Test }
    "test-coverage" { Invoke-TestCoverage }
    "lint" { Invoke-Lint }
    "fmt" { Invoke-Format }
    "vet" { Invoke-Vet }
    "clean" { Invoke-Clean }
    "deps" { Invoke-Deps }
    "tidy" { Invoke-Tidy }
    "check" { Invoke-Check }
    "ci" { Invoke-CI }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Show-Help
        exit 1
    }
}
