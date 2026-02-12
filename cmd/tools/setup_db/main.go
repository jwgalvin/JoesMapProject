package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "./data/geopulse.db"
	migrationFile := "./migrations/000001_create_events_table.up.sql"
	seedFile := "./scripts/seed_data.sql"

	fmt.Println("🗄️  GeoPulse Database Setup")
	fmt.Println("============================")
	fmt.Println()

	// Create data directory
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Remove existing database
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("⚠️  Removing existing database...")
		if err := os.Remove(dbPath); err != nil {
			log.Fatalf("Failed to remove existing database: %v", err)
		}
	}

	// Open database connection
	fmt.Println("📝 Creating database...")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Read and execute migration
	fmt.Println("🔧 Running migrations...")
	schema, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("✅ Schema created successfully!")

	// Check if seed file exists
	if _, err := os.Stat(seedFile); err == nil {
		fmt.Println("🌍 Seeding database...")
		seedData, err := os.ReadFile(seedFile)
		if err != nil {
			log.Fatalf("Failed to read seed file: %v", err)
		}

		if _, err := db.Exec(string(seedData)); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}

		// Get count of inserted events
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
		if err != nil {
			log.Fatalf("Failed to count events: %v", err)
		}

		fmt.Printf("✅ Seeded %d events successfully!\n", count)
	} else {
		fmt.Println("⚠️  Seed file not found - skipping seed data")
	}

	fmt.Println()
	fmt.Println("============================")
	fmt.Println("✅ Database setup complete!")
	fmt.Println()
	fmt.Printf("Database location: %s\n", dbPath)
	fmt.Println()
	fmt.Println("Try a query:")
	fmt.Println("  go run ./cmd/tools/query_db")

	// Show sample data
	fmt.Println()
	fmt.Println("Sample events:")
	rows, err := db.Query("SELECT location_name, magnitude_value, event_time FROM events ORDER BY magnitude_value DESC LIMIT 5")
	if err == nil {
		defer rows.Close()
		fmt.Println()
		for rows.Next() {
			var location string
			var mag float64
			var eventTime string
			if err := rows.Scan(&location, &mag, &eventTime); err == nil {
				fmt.Printf("  M%.1f - %s (%s)\n", mag, location, eventTime[:10])
			}
		}
	}
}
