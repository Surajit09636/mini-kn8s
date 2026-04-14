package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"mini-k8s/worker/handlers"
)

func init(){
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main(){
	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/run", handlers.RunHandler)

	fmt.Println("Worker running on port:", port)

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal("Worker failed:", err)
	}
}