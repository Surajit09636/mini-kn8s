package handlers

import (
	"encoding/json"
	"net/http"
	"mini-k8s/master/Schedular"
)

type RegisterPayload struct {
	URL string `json:"url"`
}

func RegisterNetworkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload RegisterPayload 
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if payload.URL == "" {
		http.Error(w, "workerURL is required", http.StatusBadRequest)
		return
	}

	schedular.RegisterWorker(payload.URL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Worker registered successfully inside master node pool"}`))
}