package handelers

import (
	"encoding/json"
	"net/http"
	
	jwtutils "mini-k8s/pkg/middleware"
)

type verifyResponse struct {
	Message string `json:"message"`
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
}

func VerifyHandeler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the user claims automatically injected by the middleware
	claims, ok := r.Context().Value(jwtutils.UserContextKey).(*jwtutils.JWTClaims)
	if !ok {
		http.Error(w, "failed to get user context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	_ = json.NewEncoder(w).Encode(verifyResponse{
		Message: "Token is valid. User is authenticated.",
		UserID:  claims.UserID,
		Email:   claims.Email,
	})
}
