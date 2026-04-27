package dao

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func getDb(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return db, nil
}

// ValidateConnection verifies that the configured database is reachable.
func ValidateConnection(connStr string) error {
	db, err := getDb(connStr)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}(db)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
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

// Duty represents a person on duty
type Duty struct {
	ID           int64
	DutyID       string
	LastDutyDate *time.Time
}

// GetAllDuties retrieves all duty records from the database
func GetAllDuties(connStr string) ([]Duty, error) {
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

	rows, err := db.Query("SELECT id, duty_id, last_duty_date FROM duties ORDER BY duty_id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var duties []Duty
	for rows.Next() {
		var duty Duty
		if err := rows.Scan(&duty.ID, &duty.DutyID, &duty.LastDutyDate); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		duties = append(duties, duty)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return duties, nil
}

// UpdateDutyDate updates the last_duty_date for a duty record
func UpdateDutyDate(connStr string, dutyID int64, date time.Time) error {
	db, err := getDb(connStr)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}(db)

	_, err = db.Exec("UPDATE duties SET last_duty_date = $1 WHERE id = $2", date, dutyID)
	if err != nil {
		return fmt.Errorf("failed to update duty date: %w", err)
	}
	return nil
}

// ReassignDutyDate clears the date from anyone assigned for the same day and assigns it to dutyID.
func ReassignDutyDate(connStr string, dutyID int64, date time.Time) error {
	db, err := getDb(connStr)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}(db)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				fmt.Printf("failed to rollback transaction: %v", rollbackErr)
			}
		}
	}()

	_, err = tx.Exec("UPDATE duties SET last_duty_date = NULL WHERE last_duty_date = $1", date)
	if err != nil {
		return fmt.Errorf("failed to clear duty dates: %w", err)
	}

	result, err := tx.Exec("UPDATE duties SET last_duty_date = $1 WHERE id = $2", date, dutyID)
	if err != nil {
		return fmt.Errorf("failed to update duty date: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		err = fmt.Errorf("duty record %d not found", dutyID)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
