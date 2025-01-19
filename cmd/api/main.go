package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	if err := db.InitSchema(context.Background()); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	handler := Handle(db)

	http.HandleFunc("/process", handler.ProcessHandler)
	http.HandleFunc("/status/", handler.StatusHandler)

	log.Println("server starting on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
