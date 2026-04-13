package handelers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"mini-k8s/auth-service/database"
	jwtutils "mini-k8s/pkg/middleware"
	"mini-k8s/auth-service/models"
	passwordutils "mini-k8s/auth-service/utils"
	"gorm.io/gorm"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Message string `json:"message"`
	Token string `json:"token,omitempty"`
	UserID uint `json:"user_id,omitempty"`
}

func LoginHandeler(w http.ResponseWriter, r *http.Request){
	//only allow post request
	if r.Method != http.MethodPost{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if database.DB == nil {
		http.Error(w, "database is not initialized", http.StatusInternalServerError)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid Request Body", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username an password are required", http.StatusBadRequest)
		return
	}

	//fetch the user from the database
	var user models.User
	err := database.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound){
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, "failed to fetch user", http.StatusInternalServerError)
			return
		}
	}

	// validate the password
	if isValid := passwordutils.CheckPasswordHash(user.Password, req.Password); !isValid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	//generate and signin the JWT Token
	token, err := jwtutils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	// send the success response with token
	log.Printf("✅ User Logged In! User ID: %d, Email: %s", user.ID, user.Email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// send back user_id and token as a bearer token payload
	_ = json.NewEncoder(w).Encode(loginResponse{
		Message: "Login successful",
		Token: token,
		UserID: user.ID,
	})
}