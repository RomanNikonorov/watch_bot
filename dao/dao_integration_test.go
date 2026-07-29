//go:build integration

package dao

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDutyRotationSerializesConcurrentTransactions(t *testing.T) {
	connectionString := os.Getenv("TEST_CONNECTION_STR")
	if connectionString == "" {
		t.Skip("TEST_CONNECTION_STR is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	adminStore, err := Open(ctx, connectionString)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer adminStore.Close()

	schema := fmt.Sprintf("watch_bot_test_%d", time.Now().UnixNano())
	if _, err := adminStore.db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	defer func() {
		_, _ = adminStore.db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	}()

	store, err := Open(ctx, connectionStringWithSearchPath(connectionString, schema))
	if err != nil {
		t.Fatalf("open schema-scoped database: %v", err)
	}
	defer store.Close()
	store.db.SetMaxOpenConns(2)

	for _, migrationPath := range []string{
		"../migrations/001_initial_schema.sql",
		"../migrations/002_unique_duty_date.sql",
	} {
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", migrationPath, err)
		}
		if _, err := store.db.ExecContext(ctx, string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", migrationPath, err)
		}
	}
	if _, err := store.db.ExecContext(
		ctx,
		"INSERT INTO duties (duty_id) VALUES ('alice'), ('bob')",
	); err != nil {
		t.Fatalf("insert duties: %v", err)
	}

	var active int32
	var maximum int32
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, 2)

	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errorsChannel <- store.WithDutyRotation(ctx, func(_ []Duty, _ *DutyRotation) error {
				current := atomic.AddInt32(&active, 1)
				for {
					observed := atomic.LoadInt32(&maximum)
					if current <= observed || atomic.CompareAndSwapInt32(&maximum, observed, current) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&active, -1)
				return nil
			})
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("duty transaction failed: %v", err)
		}
	}
	if maximum != 1 {
		t.Fatalf("expected serialized transactions, observed %d concurrent callbacks", maximum)
	}
}

func connectionStringWithSearchPath(connectionString, schema string) string {
	if strings.Contains(connectionString, "://") {
		parsed, err := url.Parse(connectionString)
		if err == nil {
			query := parsed.Query()
			query.Set("search_path", schema)
			parsed.RawQuery = query.Encode()
			return parsed.String()
		}
	}
	return connectionString + " search_path=" + schema
}
