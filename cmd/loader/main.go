package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	http.HandleFunc("/load-test", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				response, err := http.Get("http://localhost:8081/process")

				if err != nil {
					log.Printf("request failed: %v", err)
					return
				}

				body, err := io.ReadAll(response.Body)

				if err != nil {
					log.Printf("read response failed: %v", err)
					return
				}

				log.Println(string(body))

				defer response.Body.Close()

				mu.Lock()
				successCount++
				mu.Unlock()
			}()
		}

		wg.Wait()
		duration := time.Since(start)

		log.Printf("load test completed in %v. successful requests: %d", duration, successCount)
		fmt.Fprintf(w, "load test completed in %v. successful requests: %d", duration, successCount)
	})

	log.Printf("loader starting on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
