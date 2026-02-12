package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "./data/geopulse.db"

	// Check if database exists
	if _, err := os.Stat(dbPath); err != nil {
		log.Fatalf("❌ Database not found at %s\nRun: go run ./cmd/tools/setup_db", dbPath)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// If query provided as argument, run it
	if len(os.Args) > 1 {
		query := strings.Join(os.Args[1:], " ")
		runCustomQuery(db, query)
		return
	}

	// Otherwise show menu of common queries
	showMenu(db)
}

func showMenu(db *sql.DB) {
	fmt.Println("🗄️  GeoPulse Database Query Tool")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Common Queries:")
	fmt.Println()

	// Count all events
	var count int
	db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	fmt.Printf("1. Total events: %d\n", count)
	fmt.Println()

	// Largest magnitude events
	fmt.Println("2. Top 10 Largest Magnitude Events:")
	rows, err := db.Query(`
		SELECT location_name, magnitude_value, magnitude_scale, event_time 
		FROM events 
		ORDER BY magnitude_value DESC 
		LIMIT 10
	`)
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var location, scale, eventTime string
		var mag float64
		if err := rows.Scan(&location, &mag, &scale, &eventTime); err == nil {
			fmt.Printf("   M%.1f (%s) - %s (%s)\n", mag, scale, location, eventTime[:10])
		}
	}
	fmt.Println()

	// Recent events
	fmt.Println("3. Most Recent Events:")
	rows, err = db.Query(`
		SELECT location_name, magnitude_value, event_time 
		FROM events 
		ORDER BY event_time DESC 
		LIMIT 5
	`)
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var location, eventTime string
		var mag float64
		if err := rows.Scan(&location, &mag, &eventTime); err == nil {
			fmt.Printf("   M%.1f - %s (%s)\n", mag, location, eventTime[:19])
		}
	}
	fmt.Println()

	// Event type distribution
	fmt.Println("4. Events by Type:")
	rows, err = db.Query(`
		SELECT event_type, COUNT(*) as count 
		FROM events 
		GROUP BY event_type 
		ORDER BY count DESC
	`)
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err == nil {
			fmt.Printf("   %s: %d\n", eventType, count)
		}
	}
	fmt.Println()

	// Depth statistics
	fmt.Println("5. Depth Statistics:")
	var avgDepth, minDepth, maxDepth float64
	db.QueryRow("SELECT AVG(depth_km), MIN(depth_km), MAX(depth_km) FROM events").Scan(&avgDepth, &minDepth, &maxDepth)
	fmt.Printf("   Average: %.1f km\n", avgDepth)
	fmt.Printf("   Shallowest: %.1f km\n", minDepth)
	fmt.Printf("   Deepest: %.1f km\n", maxDepth)
	fmt.Println()

	// Usage instructions
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Run custom queries:")
	fmt.Println(`  go run ./cmd/tools/query_db "SELECT * FROM events WHERE magnitude_value >= 7.0"`)
	fmt.Println()
}

func runCustomQuery(db *sql.DB, query string) {
	fmt.Printf("Running query: %s\n\n", query)

	// Check if it's a SELECT query
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT") {
		// For non-SELECT queries (INSERT, UPDATE, DELETE)
		result, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Query error: %v", err)
		}
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✅ Query executed successfully. Rows affected: %d\n", rowsAffected)
		return
	}

	// For SELECT queries
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}

	// Print header
	for i, col := range columns {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Print(col)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", len(columns)*20))

	// Print rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		for i, val := range values {
			if i > 0 {
				fmt.Print(" | ")
			}
			if val == nil {
				fmt.Print("NULL")
			} else {
				fmt.Printf("%v", val)
			}
		}
		fmt.Println()
		rowCount++
	}

	fmt.Println()
	fmt.Printf("(%d rows)\n", rowCount)
}
