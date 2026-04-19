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
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: Could not load .env file from ../.env (this is fine if running via Docker or if env vars are already set)")
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