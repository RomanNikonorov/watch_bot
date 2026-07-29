package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Store owns the shared PostgreSQL connection pool used by the application.
type Store struct {
	db *sql.DB
}

// Open creates and validates a PostgreSQL connection pool.
func Open(ctx context.Context, connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	store := &Store{db: db}
	if err := store.Ping(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("database store is not initialized")
	}
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// GetUnusualDays retrieves calendar exceptions starting at currentDate.
func (s *Store) GetUnusualDays(ctx context.Context, currentDate time.Time) ([]time.Time, error) {
	rows, err := s.db.QueryContext(
		ctx,
		"SELECT unusual_date FROM unusual_days WHERE unusual_date >= $1 ORDER BY unusual_date",
		currentDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query unusual days: %w", err)
	}
	defer rows.Close()

	var days []time.Time
	for rows.Next() {
		var day time.Time
		if err := rows.Scan(&day); err != nil {
			return nil, fmt.Errorf("failed to scan unusual day: %w", err)
		}
		days = append(days, day)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate unusual days: %w", err)
	}
	return days, nil
}

// Duty represents a person participating in the duty rotation.
type Duty struct {
	ID           int64
	DutyID       string
	LastDutyDate *time.Time
}

// DutyRotation is a transaction holding row locks for all duty records.
type DutyRotation struct {
	tx *sql.Tx
}

// WithDutyRotation locks the duty set, invokes fn, and commits its updates atomically.
func (s *Store) WithDutyRotation(
	ctx context.Context,
	fn func([]Duty, *DutyRotation) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin duty transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rows, err := tx.QueryContext(
		ctx,
		"SELECT id, duty_id, last_duty_date FROM duties ORDER BY duty_id ASC FOR UPDATE",
	)
	if err != nil {
		return fmt.Errorf("failed to query duties for update: %w", err)
	}

	var duties []Duty
	for rows.Next() {
		var duty Duty
		if err := rows.Scan(&duty.ID, &duty.DutyID, &duty.LastDutyDate); err != nil {
			_ = rows.Close()
			return fmt.Errorf("failed to scan duty: %w", err)
		}
		duties = append(duties, duty)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("failed to close duty rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate duties: %w", err)
	}

	if err := fn(duties, &DutyRotation{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit duty transaction: %w", err)
	}
	return nil
}

// SetDutyDate assigns date to one duty record.
func (r *DutyRotation) SetDutyDate(ctx context.Context, dutyID int64, date time.Time) error {
	result, err := r.tx.ExecContext(
		ctx,
		"UPDATE duties SET last_duty_date = $1 WHERE id = $2",
		date,
		dutyID,
	)
	if err != nil {
		return fmt.Errorf("failed to update duty date: %w", err)
	}
	return requireAffectedDuty(result, dutyID)
}

// ReassignDutyDate clears date from its current owner and assigns it to dutyID.
func (r *DutyRotation) ReassignDutyDate(ctx context.Context, dutyID int64, date time.Time) error {
	if _, err := r.tx.ExecContext(
		ctx,
		"UPDATE duties SET last_duty_date = NULL WHERE last_duty_date = $1",
		date,
	); err != nil {
		return fmt.Errorf("failed to clear duty dates: %w", err)
	}

	result, err := r.tx.ExecContext(
		ctx,
		"UPDATE duties SET last_duty_date = $1 WHERE id = $2",
		date,
		dutyID,
	)
	if err != nil {
		return fmt.Errorf("failed to update duty date: %w", err)
	}
	return requireAffectedDuty(result, dutyID)
}

func requireAffectedDuty(result sql.Result, dutyID int64) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected duty rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("duty record %d not found", dutyID)
	}
	return nil
}
