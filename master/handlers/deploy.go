package handlers
import (
	"bytes"
	"io"
	"os"
	"encoding/json"
	"log"
	"net/http"
	jwtutils "mini-k8s/pkg/middleware"
)

// DeployPayload represents the JSON sent by the user
type DeployPayload struct {
	Image string `json:"image"`
	Replicas int `json:"replicas"`
}

func DeployHandler(w http.ResponseWriter, r *http.Request) {
	// only allow post requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the user claims automatically injected by the JWT middleware
	claims, ok := r.Context().Value(jwtutils.UserContextKey).(*jwtutils.JWTClaims)
	if !ok {
		http.Error(w, "failed to get user context", http.StatusInternalServerError)
		return
	}

	// Read the payload(Image, Replicas)
	var payload DeployPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Just print it 
	log.Printf("Deploy Request Received")
	log.Printf("User ID: %d (Email: %s)", claims.UserID, claims.Email)
	log.Printf("Requested Image: %s", payload.Image)
	log.Printf("Requsted Replicas: %d", payload.Replicas)

	// Master Worker communication
	// Default to localhost:8082 if WORKER_URL is not provided in env

	workerURL := os.Getenv("WORKER_URL")
	if workerURL == "" {
		workerURL = "http://localhost:8082"
	}

	// step 1 create the payload for the worker
	workerPayload := map[string]string{
		"image": payload.Image,
	}

	// step:2 Convert to JSON
	jsonData, err := json.Marshal(workerPayload)
	if err != nil {
		http.Error(w, "Failed to marshal worker payload", http.StatusInternalServerError)
		return
	}

	// send POST request to the worker's /run endpoint
	resp, err := http.Post(workerURL+"/run", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to connect worker: %v", err)
		http.Error(w, "Failed to forward to worker", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// step 4: Read and log the response from the worker
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response from worker: %s", string(body))

	// send a success response back to the original client
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "Deployment request validated by master and forward to worker"}`)) 
}