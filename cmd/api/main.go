package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Response struct {
	ID        uuid.UUID       `json:"id"`
	Data      json.RawMessage `json:"data"`
	Status    string          `json:"status"`
	Timestamp time.Time       `json:"timestamp"`
}

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

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	if err := db.InitSchema(context.Background()); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := db.CreatePendingResponse(ctx)
		if err != nil {
			http.Error(w, "Failed to create pending response", http.StatusInternalServerError)
			return
		}

		go func() {
			c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8082/call", nil)
			if err != nil {
				log.Printf("dial error: %v", err)
				return
			}
			defer c.Close()

			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			if err := db.SaveResponse(context.Background(), id, json.RawMessage(message)); err != nil {
				log.Printf("failed to save response: %v", err)
			}

			log.Printf("processing completed for id %s", id)
		}()

		resp := map[string]string{
			"id":      id.String(),
			"message": "started processing",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/status/"):]
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		resp, err := db.GetResponse(r.Context(), id)
		if err != nil {
			http.Error(w, "Response not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("server starting on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
