package handlers
import (
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

	// send a success response back
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "Deploy request received by master"}`))
}