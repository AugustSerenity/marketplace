package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	dto "github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	"github.com/AugustSerenity/marketplace/internal/middleware"
	"github.com/AugustSerenity/marketplace/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	service Service
	secret  string
}

func New(s Service, secret string) *Handler {
	return &Handler{
		service: s,
		secret:  secret,
	}
}

func (h *Handler) Route() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /auth-register", h.UserRegistration)
	router.HandleFunc("POST /auth-login", h.LoginUser)
	router.Handle("POST /ads", middleware.AuthMiddleware(h.secret)(http.HandlerFunc(h.CreateAd)))

	return router
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	token, err := h.service.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := auth.LoginResponse{
		Token: token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Регистрирую пользователя
func (h *Handler) UserRegistration(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var userData dto.RegistrationRequest
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := model.User{
		Login:        userData.Login,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	err = h.service.RegistrationUser(r.Context(), &user)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	resp := dto.RegistrationResponse{
		ID:    user.ID,
		Login: user.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDVal := r.Context().Value("userID")
	userID, ok := userIDVal.(uint64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ad.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	newAd := &model.Ad{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		AuthorID:    int64(userID),
		CreatedAt:   time.Now(),
	}

	if err := h.service.CreateAd(r.Context(), newAd); err != nil {
		http.Error(w, "Failed to create ad", http.StatusInternalServerError)
		return
	}

	resp := ad.Response{
		ID:          newAd.ID,
		Title:       newAd.Title,
		Description: newAd.Description,
		ImageURL:    newAd.ImageURL,
		Price:       newAd.Price,
		AuthorID:    newAd.AuthorID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
