package handlers
import (
	"bytes"
	"io"
	"os"
	"encoding/json"
	"log"
	"net/http"
	jwtutils "mini-k8s/pkg/middleware"
	"mini-k8s/master/Schedular"
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

	//Safety check: ensure at least 1 replicas is spun up if not configured
	if payload.Replicas <= 0 {
		payload.Replicas = 1
	}

	// Just print it 
	log.Printf("Deploy Request Received")
	log.Printf("User ID: %d (Email: %s)", claims.UserID, claims.Email)
	log.Printf("Requested Image: %s", payload.Image)
	log.Printf("Requsted Replicas: %d", payload.Replicas)

	workerPayload := map[string]string{
		"image": payload.Image,
	}

	jsonData, err := json.Marshal(workerPayload)
	if err != nil {
		http.Error(w, "Failed to marshal worker payload", http.StatusInternalServerError)
		return
	}

	//Introduce a for loop based on the number of replicas
	for i := 0; i < payload.Replicas; i++ {
		// Ask the schedular for the next server in the loop
		workerURL := schedular.GetNextWorker()

		// Fallback local setting if the array in the schedular is empty
		if workerURL == "" {
			workerURL = os.Getenv("WORKER_URL")
			if workerURL == "" {
				workerURL = "http://localhost:8082"
			}
		}

		log.Printf("deploying Replica %d/%d to worker: %s", i+1, payload.Replicas, workerURL)

		// send POST request sequentially or inside a go-routine
		resp, err := http.Post(workerURL+"/run", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("failed to connect to worker %s: %v", workerURL, err)
			continue // if one replicas failed on a node, continue trying to deploy the others
		}

		body, _ := io.ReadAll(resp.Body)
		log.Printf("Response from worker %s on replica %d: %s", workerURL, i+1, string(body))
		resp.Body.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "Deployment request validate by master and forwarded to worker nodes"}`))
}