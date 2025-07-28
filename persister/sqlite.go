package persister

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite"
)

var _ Persister[Element] = (*SQLitePersister[Element])(nil)

type SQLitePersister[T Element] struct {
	filepath string
	db       *sql.DB
}

func NewSQLitePersister[T Element](filepath string) (*SQLitePersister[T], error) {
	u := url.URL{
		Scheme: "file",
		Path:   filepath,
	}

	q := u.Query() // Get a copy
	q.Set("_journal_mode", "WAL")
	q.Set("_busy_timeout", "5000")
	u.RawQuery = q.Encode() // Save back to URL

	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS elements (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLitePersister[T]{
		filepath: u.String(),
		db:       db,
	}, nil
}

// Delete implements Persister.
func (s *SQLitePersister[T]) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM elements WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item with id %s not found", id)
	}

	return nil
}

// Get implements Persister.
func (s *SQLitePersister[T]) Get(ctx context.Context, id string) (T, error) {
	var zero T
	var dataStr string

	query := `SELECT data FROM elements WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&dataStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return zero, fmt.Errorf("item with id %s not found", id)
		}
		return zero, fmt.Errorf("failed to get item: %w", err)
	}

	var item T
	if err := json.Unmarshal([]byte(dataStr), &item); err != nil {
		return zero, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return item, nil
}

// List implements Persister.
func (s *SQLitePersister[T]) List(ctx context.Context, offset int, limit int) ([]T, error) {
	query := `SELECT data FROM elements ORDER BY id LIMIT ? OFFSET ?`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var item T
		if err := json.Unmarshal([]byte(dataStr), &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return items, nil
}

// Close closes the database connection
func (s *SQLitePersister[T]) Close() error {
	return s.db.Close()
}

// Save implements Persister.
func (s *SQLitePersister[T]) Save(ctx context.Context, item T) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	query := `INSERT OR REPLACE INTO elements (id, data) VALUES (?, ?)`
	_, err = s.db.ExecContext(ctx, query, item.Id(), string(data))
	if err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}

	return nil
}
