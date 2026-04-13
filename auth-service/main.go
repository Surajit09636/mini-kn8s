package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"mini-k8s/auth-service/database"
	"mini-k8s/auth-service/Handelers"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if database.HasConfig() {
		if err := database.ConnectDB(); err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
	} else {
		log.Println("Database configuration not set, starting without a DB connection")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("auth-service is running"))
	})

	// all routers
	mux.HandleFunc("POST /signup", handelers.SignupHandeler)
	mux.HandleFunc("POST /login", handelers.LoginHandeler)
	
	// Protected route
	mux.HandleFunc("GET /verify", jwtutils.JWTMiddleware(handelers.VerifyHandeler))

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}

}
