package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	db *DB
}

func Handle(db *DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ProcessHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := h.db.CreatePendingResponse(ctx)
	if err != nil {
		http.Error(w, "failed to create pending response", http.StatusInternalServerError)
		return
	}

	go h.processWebSocket(id)

	resp := map[string]string{
		"id":      id.String(),
		"message": "started processing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/status/"):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	resp, err := h.db.GetResponse(r.Context(), id)
	if err != nil {
		http.Error(w, "response not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) processWebSocket(id uuid.UUID) {
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

	if err := h.db.SaveResponse(context.Background(), id, json.RawMessage(message)); err != nil {
		log.Printf("failed to save response: %v", err)
	}

	log.Printf("processing completed for id %s", id)
}
