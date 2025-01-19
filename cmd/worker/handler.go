package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

func fetchData(id string) (string, error) {

	// this is a public api to mimic a concurrent task
	baseUrl := "https://jsonplaceholder.typicode.com/todos"
	if id != "" {
		baseUrl = "https://jsonplaceholder.typicode.com/todos/" + id
	}

	response, err := http.Get(baseUrl)
	if err != nil {
		return "", fmt.Errorf("error making external request: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {

	//simulate delay at random intervals
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading to web socket protocol: %v", err)
		return
	}
	defer ws.Close()

	atomic.AddInt64(&s.activeConnections, 1)
	defer atomic.AddInt64(&s.activeConnections, -1)

	go func() {

		response, err := fetchData(strconv.Itoa(rand.Intn(200) + 1))
		if err != nil {
			log.Printf("write error: %v", err)
		}

		err = ws.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			log.Printf("write error: %v", err)
		}
	}()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}
