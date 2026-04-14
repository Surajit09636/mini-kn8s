package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"mini-k8s/master/handlers"
	jwtutils "mini-k8s/pkg/middleware"
)

func init() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found")
		}
	}
}

func main() {
	// The Master node can run on a different port than Auth, let's say 8081
	port := os.Getenv("MASTER_PORT")
	if port == "" {
		port = "8081" 
	}

	mux := http.NewServeMux()

	// Public health check route
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("Master node is running"))
	})

	// 🚀 Protected /deploy route using your shared pkg/middleware
	mux.HandleFunc("POST /deploy", jwtutils.JWTMiddleware(handlers.DeployHandler))

	log.Printf("Starting Master server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Master Server failed to start:", err)
	}
}
