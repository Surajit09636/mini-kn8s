package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	schedular "mini-k8s/master/Schedular"
	"mini-k8s/master/database"
	"mini-k8s/master/models"
	jwtutils "mini-k8s/pkg/middleware"
)

type DeployPayload struct {
	Image    string `json:"image"`
	Replicas int    `json:"replicas"`
}

type WorkerResponse struct {
	Message     string `json:"message"`
	ContainerID string `json:"container_id"`
}

func DeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(jwtutils.UserContextKey).(*jwtutils.JWTClaims)
	if !ok {
		http.Error(w, "failed to get user context", http.StatusInternalServerError)
		return
	}

	var payload DeployPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	if payload.Replicas <= 0 {
		payload.Replicas = 1
	}

	// 1. CREATE DEPLOYMENT RECORD IN DB
	deployment := models.Deployment{
		UserID:   claims.UserID,
		Image:    payload.Image,
		Replicas: payload.Replicas,
		Status:   "running",
	}
	if err := database.DB.Create(&deployment).Error; err != nil {
		http.Error(w, "Failed to save deployment", http.StatusInternalServerError)
		return
	}

	workerPayload := map[string]string{"image": payload.Image}
	jsonData, _ := json.Marshal(workerPayload)

	for i := 0; i < payload.Replicas; i++ {
		workerURL := schedular.GetNextWorker()
		if workerURL == "" {
			workerURL = "http://localhost:8082"
		} // Fallback

		resp, err := http.Post(workerURL+"/run", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("failed to connect to worker %s: %v", workerURL, err)
			continue
		}

		// 2. PARSE JSON FROM WORKER AND SAVE THE POD ID
		var workerResp WorkerResponse
		if err := json.NewDecoder(resp.Body).Decode(&workerResp); err == nil {
			pod := models.Pod{
				DeploymentID: deployment.ID,
				WorkerURL:    workerURL,
				ContainerID:  workerResp.ContainerID,
				Status:       "running",
			}
			database.DB.Create(&pod)
		}
		resp.Body.Close()

		schedular.AssignTask(workerURL, payload.Image)
	}

	// 3. RETURN DEPLOYMENT ID TO THE USER
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Deployment accepted",
		"deployment_id": deployment.ID,
	})
}
