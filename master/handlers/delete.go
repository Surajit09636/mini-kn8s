package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"mini-k8s/master/Schedular"
	"mini-k8s/master/database"
	"mini-k8s/master/models"
	jwtutils "mini-k8s/pkg/middleware"
)

type teardownPayload struct {
	ContainerID string `json:"container_id"`
}

func DeleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(jwtutils.UserContextKey).(*jwtutils.JWTClaims)
	if !ok {
		http.Error(w, "Failed to get user context", http.StatusInternalServerError)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid Deployment ID", http.StatusBadRequest)
		return
	}

	var deployment models.Deployment
	err = database.DB.Preload("Pods").Where("id = ? AND user_id = ?", uint(id), claims.UserID).First(&deployment).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "failed to load deployment", http.StatusInternalServerError)
		return
	}

	failed := make([]string, 0)

	for _, pod := range deployment.Pods {
		body, _ := json.Marshal(teardownPayload{ContainerID: pod.ContainerID})
		resp, err := http.Post(pod.WorkerURL+"/teardown", "application/json", bytes.NewBuffer(body))
		if err != nil {
			failed = append(failed, pod.ContainerID)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			failed = append(failed, pod.ContainerID)
			continue
		}

		schedular.RemoveTask(pod.WorkerURL, deployment.Image)
	}

	if len(failed) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]any{
			"message":           "some containers could not be removed; retry delete",
			"failed_containers": failed,
		})
		return
	}

	if err := database.DB.Where("deployment_id = ?", deployment.ID).Delete(&models.Pod{}).Error; err != nil {
		http.Error(w, "failed to delete pods", http.StatusInternalServerError)
		return
	}

	if err := database.DB.Delete(&deployment).Error; err != nil {
		http.Error(w, "failed to delete deployment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":       "deployment deleted successfully",
		"deployment_id": deployment.ID,
	})
}
