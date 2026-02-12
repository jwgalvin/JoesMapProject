# Seed the GeoPulse database with realistic earthquake data
# Run this after init_db.ps1

$ErrorActionPreference = "Stop"

$DBPath = ".\data\geopulse.db"
$SeedFile = ".\scripts\seed_data.sql"

Write-Host "🌍 Seeding GeoPulse Database..." -ForegroundColor Cyan

# Check if database exists
if (-not (Test-Path $DBPath)) {
    Write-Host "❌ Error: Database not found at $DBPath" -ForegroundColor Red
    Write-Host "Run .\scripts\init_db.ps1 first to create the database" -ForegroundColor Yellow
    exit 1
}

# Check if seed file exists
if (-not (Test-Path $SeedFile)) {
    Write-Host "❌ Error: Seed file not found at $SeedFile" -ForegroundColor Red
    exit 1
}

# Check if sqlite3 CLI is available
$sqliteInstalled = Get-Command sqlite3 -ErrorAction SilentlyContinue
if (-not $sqliteInstalled) {
    Write-Host "❌ Error: sqlite3 CLI not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installation Options:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Option 1 - Direct Download (Recommended for Windows):" -ForegroundColor Cyan
    Write-Host "  1. Download from: https://www.sqlite.org/download.html" -ForegroundColor White
    Write-Host "  2. Get: sqlite-tools-win32-x86-*.zip" -ForegroundColor White
    Write-Host "  3. Extract sqlite3.exe to a folder in your PATH" -ForegroundColor White
    Write-Host "     (e.g., C:\Program Files\SQLite\ or add current dir to PATH)" -ForegroundColor White
    Write-Host ""
    Write-Host "Option 2 - Using Scoop (if you have it installed):" -ForegroundColor Cyan
    Write-Host "  scoop install sqlite" -ForegroundColor White
    Write-Host ""
    exit 1
}

# Import seed data
Write-Host "Importing seed data..." -ForegroundColor Green
Get-Content $SeedFile | sqlite3 $DBPath

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Database seeded successfully!" -ForegroundColor Green
    Write-Host ""
    
    # Show summary
    Write-Host "Database Summary:" -ForegroundColor Cyan
    $count = sqlite3 $DBPath "SELECT COUNT(*) FROM events;"
    Write-Host "  Total events: $count" -ForegroundColor White
    
    $largestMag = sqlite3 $DBPath "SELECT MAX(magnitude_value) FROM events;"
    Write-Host "  Largest magnitude: $largestMag" -ForegroundColor White
    
    $eventTypes = sqlite3 $DBPath "SELECT COUNT(DISTINCT event_type) FROM events;"
    Write-Host "  Event types: $eventTypes" -ForegroundColor White
    
    Write-Host ""
    Write-Host "Sample queries:" -ForegroundColor Yellow
    Write-Host "  sqlite3 $DBPath 'SELECT * FROM events LIMIT 5;'" -ForegroundColor Gray
    Write-Host "  sqlite3 $DBPath 'SELECT * FROM events WHERE magnitude_value >= 7.0 ORDER BY magnitude_value DESC;'" -ForegroundColor Gray
    Write-Host "  sqlite3 $DBPath 'SELECT event_type, COUNT(*) as count FROM events GROUP BY event_type;'" -ForegroundColor Gray
} else {
    Write-Host "❌ Seeding failed!" -ForegroundColor Red
    exit 1
}
