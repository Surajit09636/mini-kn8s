package handlers

import (
	"encoding/json"
	"mini-k8s/master/database"
	"mini-k8s/master/models"
	jwtutils "mini-k8s/pkg/middleware"
	"net/http"
)

func GetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(jwtutils.UserContextKey).(*jwtutils.JWTClaims)
	if !ok {
		http.Error(w, "failed to get user context", http.StatusInternalServerError)
		return
	}

	// Fetch Deployments and cleanly preload their associated Pods automatically
	var deployments []models.Deployment
	if err := database.DB.Preload("Pods").Where("user_id = ?", claims.UserID).Find(&deployments).Error; err != nil {
		http.Error(w, "failed to fetch deployments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deployments)
}
