# Initialize local SQLite database for development
# This script creates the database file and runs migrations

$ErrorActionPreference = "Stop"

$DBPath = ".\data\geopulse.db"
$DBDir = ".\data"
$MigrationsPath = ".\migrations"

Write-Host "🗄️  Initializing GeoPulse Database..." -ForegroundColor Cyan

# Create data directory if it doesn't exist
if (-not (Test-Path $DBDir)) {
    Write-Host "Creating data directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $DBDir | Out-Null
}

# Remove existing database if it exists
if (Test-Path $DBPath) {
    Write-Host "⚠️  Removing existing database..." -ForegroundColor Yellow
    Remove-Item $DBPath -Force
}

# Check if golang-migrate is installed
$migrateInstalled = Get-Command migrate -ErrorAction SilentlyContinue
if (-not $migrateInstalled) {
    Write-Host "❌ Error: golang-migrate not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installation Options:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Option 1 - Direct Download (Recommended for Windows):" -ForegroundColor Cyan
    Write-Host "  1. Download from: https://github.com/golang-migrate/migrate/releases" -ForegroundColor White
    Write-Host "  2. Get: migrate.windows-amd64.zip" -ForegroundColor White
    Write-Host "  3. Extract migrate.exe to a folder in your PATH" -ForegroundColor White
    Write-Host "     (e.g., C:\Program Files\migrate\ or add current dir to PATH)" -ForegroundColor White
    Write-Host ""
    Write-Host "Option 2 - Using Scoop (if you have it installed):" -ForegroundColor Cyan
    Write-Host "  scoop install migrate" -ForegroundColor White
    Write-Host ""
    Write-Host "Option 3 - Using Go:" -ForegroundColor Cyan
    Write-Host "  go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest" -ForegroundColor White
    Write-Host ""
    exit 1
}

# Run migrations
Write-Host "Running migrations..." -ForegroundColor Green
$databaseURL = "sqlite3://$DBPath"

# Create the database file using migrate
migrate -path $MigrationsPath -database $databaseURL up

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Database initialized successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Database location: $DBPath" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Run: .\scripts\seed_db.ps1   # Populate with sample data" -ForegroundColor White
    Write-Host "  2. Or manually insert data using SQLite CLI" -ForegroundColor White
} else {
    Write-Host "❌ Migration failed!" -ForegroundColor Red
    exit 1
}
