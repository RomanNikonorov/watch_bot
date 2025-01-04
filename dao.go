package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func readFromDatabase(connStr string) ([]Server, error) {

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
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

	var servers []Server
	for rows.Next() {
		var server Server
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
