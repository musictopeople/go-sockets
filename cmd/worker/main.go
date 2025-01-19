package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	activeConnections int64
	upgrader          websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func main() {
	server := NewServer()

	http.HandleFunc("/call", server.HandleRequest)

	log.Println("worker starting on localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
