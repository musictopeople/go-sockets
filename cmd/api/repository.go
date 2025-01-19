package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func InitDB() (*DB, error) {
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	config.MaxConns = 50
	config.MinConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) InitSchema(ctx context.Context) error {
	sql := `
    CREATE TABLE IF NOT EXISTS responses (
        id UUID PRIMARY KEY,
        data JSONB,
        status TEXT NOT NULL DEFAULT 'pending',
        timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS responses_data_idx ON responses USING GIN (data);`

	_, err := db.pool.Exec(ctx, sql)
	return err
}

func (db *DB) SaveResponse(ctx context.Context, id uuid.UUID, data json.RawMessage) error {
	sql := `UPDATE responses SET data = $2, status = 'completed' WHERE id = $1`
	_, err := db.pool.Exec(ctx, sql, id, data)
	return err
}

func (db *DB) CreatePendingResponse(ctx context.Context) (uuid.UUID, error) {
	id := uuid.New()
	sql := `INSERT INTO responses (id, status) VALUES ($1, 'pending')`
	_, err := db.pool.Exec(ctx, sql, id)
	return id, err
}

func (db *DB) GetResponse(ctx context.Context, id uuid.UUID) (*Response, error) {
	var resp Response
	sql := `SELECT id, COALESCE(data, '{}'::jsonb), status, timestamp FROM responses WHERE id = $1`

	err := db.pool.QueryRow(ctx, sql, id).Scan(&resp.ID, &resp.Data, &resp.Status, &resp.Timestamp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
