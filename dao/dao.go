package dao

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"

	"watch_bot/watch"
)

func getDb(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return db, nil
}

// GetUnusualDays retrieves the list of unusual days from the database.
func GetUnusualDays(connStr string, currentDate time.Time) ([]time.Time, error) {
	db, err := getDb(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}(db)

	rows, err := db.Query("select unusual_days.unusual_date from unusual_days where unusual_date >= $1", currentDate)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var days []time.Time
	for rows.Next() {
		var day time.Time
		if err := rows.Scan(&day); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		days = append(days, day)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return days, nil
}

// GetServers retrieves the list of servers from the database.
func GetServers(connStr string) ([]watch.Server, error) {

	db, err := getDb(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}(db)
	rows, err := db.Query("select name, url from servers")
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var servers []watch.Server
	for rows.Next() {
		var server watch.Server
		if err := rows.Scan(&server.Name, &server.URL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		fmt.Printf("Name: %s, url: %s\n", server.Name, server.URL)
		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return servers, nil
}
