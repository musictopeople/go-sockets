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

	handle := Handle(db)

	http.HandleFunc("/process", handle.Process)
	http.HandleFunc("/status/", handle.Status)

	log.Println("server starting on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
