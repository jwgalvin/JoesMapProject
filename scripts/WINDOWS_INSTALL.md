# Windows Installation Guide

Quick guide for setting up the GeoPulse database tools on Windows (without Scoop).

## Tools Needed

1. **golang-migrate** - Database migration tool
2. **sqlite3** - SQLite command-line tool

## Installation

### Method 1: Direct Downloads (Recommended)

#### Install golang-migrate

1. **Download** the latest release:
   - Go to: https://github.com/golang-migrate/migrate/releases
   - Download: `migrate.windows-amd64.zip` (or latest version)
   
2. **Extract** the ZIP file

3. **Install** the binary:
   ```powershell
   # Option A: Add to system PATH
   # 1. Create directory: C:\Program Files\migrate
   # 2. Copy migrate.exe there
   # 3. Add to PATH via System Environment Variables
   
   # Option B: Add to project (simpler)
   # 1. Create .\bin directory in your project
   mkdir bin
   # 2. Copy migrate.exe to .\bin\
   # 3. The scripts will find it if it's in PATH or current directory
   ```

4. **Verify installation**:
   ```powershell
   migrate -version
   ```

#### Install SQLite3

1. **Download** SQLite tools:
   - Go to: https://www.sqlite.org/download.html
   - Download: `sqlite-tools-win32-x86-*.zip` (under "Precompiled Binaries for Windows")

2. **Extract** the ZIP file (contains `sqlite3.exe`)

3. **Install** the binary:
   ```powershell
   # Option A: Add to system PATH
   # 1. Create directory: C:\Program Files\SQLite
   # 2. Copy sqlite3.exe there
   # 3. Add to PATH via System Environment Variables
   
   # Option B: Add to project (simpler)
   # 1. Copy sqlite3.exe to your project's bin\ directory
   mkdir bin
   # 2. Add .\bin to your PATH for this session:
   $env:PATH += ";$PWD\bin"
   ```

4. **Verify installation**:
   ```powershell
   sqlite3 -version
   ```

### Method 2: Using Go Install (for migrate only)

If you already have Go installed:

```powershell
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

This installs `migrate.exe` to your `$GOPATH\bin` directory (usually `C:\Users\<username>\go\bin`).

Make sure `%GOPATH%\bin` is in your PATH.

### Method 3: Using Scoop (if you want to install it)

Scoop is a command-line installer for Windows (like apt/brew):

```powershell
# Install Scoop first
irm get.scoop.sh | iex

# Then install tools
scoop install migrate sqlite
```

## Adding to PATH (Windows)

### Temporary (current session only):

```powershell
# Add to current PowerShell session
$env:PATH += ";C:\Program Files\migrate"
$env:PATH += ";C:\Program Files\SQLite"
```

### Permanent (system-wide):

1. Press `Win + X`, select "System"
2. Click "Advanced system settings"
3. Click "Environment Variables"
4. Under "System variables", select "Path" and click "Edit"
5. Click "New" and add your directory (e.g., `C:\Program Files\migrate`)
6. Click "OK" on all dialogs
7. Restart PowerShell

## Quick Setup

After installing both tools:

```powershell
# Initialize database
.\scripts\init_db.ps1

# Seed with data
.\scripts\seed_db.ps1

# Query database
sqlite3 data/geopulse.db "SELECT COUNT(*) FROM events;"
```

## Alternative: Manual Setup (No Extra Tools)

If you don't want to install migrate/sqlite3, you can set up the database manually with Go.

The project already includes `cmd/tools/setup_db` and `cmd/tools/query_db` directories with Go tools.

Run with:
```powershell
# Setup database
go run ./cmd/tools/setup_db

# Query database
go run ./cmd/tools/query_db
```

## Troubleshooting

### "migrate is not recognized"

- Verify migrate.exe is in your PATH
- Try running with full path: `C:\path\to\migrate.exe -version`
- Restart your terminal after adding to PATH

### "sqlite3 is not recognized"

- Verify sqlite3.exe is in your PATH
- Try running with full path: `C:\path\to\sqlite3.exe -version`
- Restart your terminal after adding to PATH

### "Permission denied"

- Run PowerShell as Administrator
- Check Windows Defender or antivirus isn't blocking the executable

### "Execution Policy" errors

If you can't run PowerShell scripts:

```powershell
# Allow scripts for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Need Help?

- **golang-migrate docs**: https://github.com/golang-migrate/migrate
- **SQLite download**: https://www.sqlite.org/download.html
- **Scoop (optional)**: https://scoop.sh
