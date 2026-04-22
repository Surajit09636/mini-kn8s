// worker/main.go (Changes going inside your main function)
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"mini-k8s/worker/handlers"
)

func init(){
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: Could not load .env file")
	}
}

func main(){
	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8082"
	}

	// 1. Calculate this specific worker's callback URL
	myURL := "http://localhost:" + port

	// 2. Define where the Master control plane lives
	masterURL := os.Getenv("MASTER_URL") // Can add to .env later
	if masterURL == "" {
		masterURL = "http://localhost:8081" 
	}

	// 3. Ping the master node to announce our presence!
	payload := map[string]string{"url": myURL}
	jsonData, _ := json.Marshal(payload)
	
	// Background Goroutine to keep pinging Master until it correctly registers
	go func() {
		for {
			resp, err := http.Post(masterURL+"/register", "application/json", bytes.NewBuffer(jsonData))
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				log.Println("✅ Successfully registered with Master Node!")
				return // Break out of the loop completely when successful
			}

			if err != nil {
				log.Printf("⚠️ Warning: Master Node offline. Retrying registration in 5 seconds...")
			} else {
				log.Printf("⚠️ Warning: Got a non-200 status from Master Node (%d). Retrying...", resp.StatusCode)
				resp.Body.Close()
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// Your existing server spin-up code...
	mux := http.NewServeMux()
	mux.HandleFunc("/run", handlers.RunHandler)

	fmt.Println("Worker ready & listening for jobs on port:", port)

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal("Worker failed:", err)
	}
}
