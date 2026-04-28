package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

type teardownRequest struct {
	ContainerID string `json:"container_id"`
}

func TeardownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req teardownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ContainerID == "" {
		http.Error(w, "container_id is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		http.Error(w, "failed to connect to docker daemon", http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	timeout := 10
	_, err = cli.ContainerStop(ctx, req.ContainerID, client.ContainerStopOptions{Timeout: &timeout})
	if err != nil && !errdefs.IsNotFound(err) {
		log.Printf("failed to stop container %s: %v", req.ContainerID, err)
		http.Error(w, "failed to stop container", http.StatusInternalServerError)
		return
	}

	_, err = cli.ContainerRemove(ctx, req.ContainerID, client.ContainerRemoveOptions{Force: true})
	if err != nil && !errdefs.IsNotFound(err) {
		log.Printf("failed to remove container %s: %v", req.ContainerID, err)
		http.Error(w, "failed to remove container", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "container teardown successfully",
		"container_id": req.ContainerID,
	})
}
