# Database Scripts

Scripts for managing the GeoPulse local development database.

## Quick Start (Easiest Method - No Extra Tools!)

**Using Go only** (no migrate or sqlite3 needed):

```powershell
# Setup database and seed data in one command
go run ./cmd/tools/setup_db

# Query the database
go run ./cmd/tools/query_db

# Or run custom queries
go run ./cmd/tools/query_db "SELECT * FROM events WHERE magnitude_value >= 7.0"
```

This is the **recommended approach** if you already have Go installed. It doesn't require any additional tools!

## Alternative: Using Migration Tools

If you prefer using dedicated database tools:

### Prerequisites

Install the required tools:

```powershell
# Using Scoop (recommended for Windows)
scoop install migrate sqlite

# Or see WINDOWS_INSTALL.md for manual installation without Scoop
```

**Don't have Scoop?** See [WINDOWS_INSTALL.md](WINDOWS_INSTALL.md) for detailed installation instructions.

```powershell
# Initialize database (creates schema)
.\scripts\init_db.ps1

# Populate with sample data
.\scripts\seed_db.ps1
```

The database will be created at `data/geopulse.db`.

## Scripts

### `init_db.ps1`

Creates a new SQLite database and runs migrations.

- Creates the `data/` directory if needed
- Removes existing database if present
- Runs all migrations from `migrations/` directory
- Creates tables and indexes

**Usage:**

```powershell
.\scripts\init_db.ps1
```

### `seed_db.ps1`

Populates the database with realistic earthquake data.

- Inserts ~30 earthquake events
- Includes major historical earthquakes (M9.1 Tohoku, M7.8 Turkey-Syria, etc.)
- Covers various magnitudes, depths, and locations
- Provides diverse data for testing queries and filters

**Usage:**

```powershell
.\scripts\seed_db.ps1
```

### `seed_data.sql`

SQL script containing INSERT statements for seed data.

You can also import this manually:

```powershell
sqlite3 data/geopulse.db < scripts/seed_data.sql
```

## Manual Database Management

### Query the database

```powershell
# Open interactive SQLite shell
sqlite3 data/geopulse.db

# Run a quick query
sqlite3 data/geopulse.db "SELECT * FROM events WHERE magnitude_value >= 7.0 ORDER BY magnitude_value DESC LIMIT 10;"

# Count events
sqlite3 data/geopulse.db "SELECT COUNT(*) FROM events;"

# View schema
sqlite3 data/geopulse.db ".schema events"
```

### Reset the database

```powershell
# Delete and recreate
Remove-Item data/geopulse.db
.\scripts\init_db.ps1
.\scripts\seed_db.ps1
```

### Migration commands

```powershell
# Check migration status
migrate -path migrations -database "sqlite3://data/geopulse.db" version

# Migrate up (apply migrations)
migrate -path migrations -database "sqlite3://data/geopulse.db" up

# Rollback one migration
migrate -path migrations -database "sqlite3://data/geopulse.db" down 1

# Force version (if migration state is corrupted)
migrate -path migrations -database "sqlite3://data/geopulse.db" force 1
```

## Database Schema

The `events` table contains:

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT | Unique event ID (primary key) |
| `event_type` | TEXT | Type: earthquake, explosion, etc. |
| `magnitude_value` | REAL | Magnitude value (-1.0 to 10.0) |
| `magnitude_scale` | TEXT | Scale: mw, ml, md, etc. |
| `latitude` | REAL | Latitude (-90 to 90) |
| `longitude` | REAL | Longitude (-180 to 180) |
| `depth_km` | REAL | Depth in kilometers |
| `event_time` | DATETIME | When the event occurred |
| `location_name` | TEXT | Human-readable location |
| `status` | TEXT | Review status (automatic/reviewed) |
| `description` | TEXT | Event description |
| `url` | TEXT | Reference URL |
| `created_at` | DATETIME | Record creation timestamp |
| `updated_at` | DATETIME | Last update timestamp |

### Indexes

- `idx_events_magnitude` - Magnitude queries
- `idx_events_time` - Time-based queries
- `idx_events_type` - Event type filtering
- `idx_events_location` - Location-based queries
- `idx_events_time_magnitude` - Composite index for time + magnitude queries

## Sample Data

The seed data includes:

- **30+ earthquake events** from around the world
- **Magnitude range:** 2.5 to 9.1
- **Depth range:** 2.9 km to 598 km (shallow to very deep)
- **Locations:** Turkey, Japan, Indonesia, Alaska, California, Mexico, Chile, Hawaii, etc.
- **Historical events:** 2011 Tohoku (M9.1), 2010 Haiti (M7.0), 2023 Turkey-Syria (M7.8)
- **Recent activity:** Small to moderate events from 2024

Perfect for testing:
- Magnitude filters
- Time range queries
- Location-based searches
- Pagination
- Event type filtering
- Depth classification

## Troubleshooting

### "migrate command not found"

Install golang-migrate:

```powershell
scoop install migrate
```

### "sqlite3 command not found"

Install SQLite:

```powershell
scoop install sqlite
```

### "Migration failed"

- Check that migration files exist in `migrations/` directory
- Ensure database path is correct
- Verify SQL syntax in migration files

### "Database locked"

Close any open connections to the database (SQLite browser, terminals with active queries, etc.)

## Next Steps

After seeding the database:

1. **Test queries manually:**
   ```powershell
   sqlite3 data/geopulse.db "SELECT * FROM events WHERE magnitude_value >= 7.0;"
   ```

2. **Run integration tests** against the local database (once implemented)

3. **Start the API server** (once implemented) and test endpoints

4. **Verify migrations** work correctly with up/down migrations
