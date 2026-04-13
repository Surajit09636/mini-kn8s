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

type signupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signupResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	UserID uint   `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Email string `json:"email,omitempty"`
}

func SignupHandeler( w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "database is not initialized", http.StatusInternalServerError)
		return
	}

	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid Request Body", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "username, email and password are required", http.StatusBadRequest)
		return
	}

	var existingUser models.User
	err := database.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		http.Error(w, "Username or Email already exists", http.StatusConflict)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound){
		http.Error(w, "failed to check existing user", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := passwordutils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username: req.Username,
		Email:	req.Email,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := jwtutils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ New User Signed Up! User ID: %d, Username: %s, Email: %s", user.ID, user.Username, user.Email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(signupResponse{
		Message: "User created successfully",
		Token:   token,
		UserID:  user.ID,
		Username: user.Username,
		Email: user.Email,
	})

}