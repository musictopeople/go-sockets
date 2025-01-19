package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID        uuid.UUID       `json:"id"`
	Data      json.RawMessage `json:"data"`
	Status    string          `json:"status"`
	Timestamp time.Time       `json:"timestamp"`
}
