package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	// Using the newly consolidated moby client
	"github.com/moby/moby/client"

	"mini-k8s/worker/models"
)

func RunHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the request from Master
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

	// 2. Initialize Docker Client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Failed to create docker client: %v", err)
		http.Error(w, "Failed to connect to Docker daemon", http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	fmt.Println("\n🚀 Pulling Docker Image:", req.Image)

	// 3. Pull the Image (e.g., pulling nginx:latest from Docker Hub)
	pullResponse, err := cli.ImagePull(ctx, req.Image, client.ImagePullOptions{})
	if err != nil {
		log.Printf("Failed to pull image %s: %v", req.Image, err)
		http.Error(w, "Failed to pull image", http.StatusInternalServerError)
		return
	}

	// Copy stream to stdout to see download progress in the terminal
	io.Copy(os.Stdout, pullResponse)
	pullResponse.Close()

	fmt.Println("\n✅ Image pulled successfully. Creating container...")

	// 4. Create the Container configuration
	createResponse, err := cli.ContainerCreate(ctx,
		client.ContainerCreateOptions{
			Image: req.Image, // Target image
		},
	)
	if err != nil {
		log.Printf("Failed to create container: %v", err)
		http.Error(w, "Failed to create container", http.StatusInternalServerError)
		return
	}

	// 5. Start the Container
	_, err = cli.ContainerStart(ctx, createResponse.ID, client.ContainerStartOptions{})
	if err != nil {
		log.Printf("Failed to start container %s: %v", createResponse.ID, err)
		http.Error(w, "Failed to start container", http.StatusInternalServerError)
		return
	}

	log.Printf("Worker successfully started container! ID: %s", createResponse.ID)

	//Return JSON to Master!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "success",
		"container_id": createResponse.ID,
	})

	// 6. Print success to worker terminal
	successMsg := fmt.Sprintf("✅ Worker successfully started container!\nContainer ID: %s", createResponse.ID)
	fmt.Println(successMsg)
}
