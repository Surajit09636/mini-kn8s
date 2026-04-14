package handlers

import(
	"encoding/json"
	"fmt"
	"net/http"
	"mini-k8s/worker/models"
)

func RunHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RunRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	if req.Image == "" {
		http.Error(w, "Image required", http.StatusBadRequest)
		return
	}

	fmt.Println("Running container for Image:", req.Image)

	w.Write([]byte("Worker received request"))
}