package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	"github.com/AugustSerenity/marketplace/internal/middleware"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  Service
	secret   string
	validate *validator.Validate
}

func New(s Service, secret string) *Handler {
	return &Handler{
		service:  s,
		secret:   secret,
		validate: validator.New(),
	}
}

func (h *Handler) Route() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /auth-register", h.UserRegistration)
	router.HandleFunc("POST /auth-login", h.LoginUser)
	router.Handle("POST /create-ads", middleware.AuthMiddleware(h.secret)(http.HandlerFunc(h.CreateAd)))
	router.Handle("GET /watch-ads", middleware.AuthMiddleware(h.secret)(http.HandlerFunc(h.GetAds)))

	return router
}

func (h *Handler) UserRegistration(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var req auth.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.service.RegisterUser(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "User with this login already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			http.Error(w, "Request Timeout", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := auth.LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	userIDVal := r.Context().Value("userID")
	userID, ok := userIDVal.(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ad.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	createdAd, err := h.service.CreateAd(r.Context(), req, userID)
	if err != nil {
		http.Error(w, "Failed to create ad: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := ad.Response{
		ID:          createdAd.ID,
		Title:       createdAd.Title,
		Description: createdAd.Description,
		ImageURL:    createdAd.ImageURL,
		Price:       createdAd.Price,
		AuthorID:    createdAd.AuthorID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetAds(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	req, err := h.service.ParseListRequest(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := int64(0)
	if userIDVal := r.Context().Value("userID"); userIDVal != nil {
		if id, ok := userIDVal.(int64); ok {
			userID = id
		}
	}

	ads, err := h.service.GetAds(r.Context(), &req, userID)
	if err != nil {
		http.Error(w, "Failed to fetch ads", http.StatusInternalServerError)
		return
	}

	var resp []ad.ListResponse
	for _, adItem := range ads {
		resp = append(resp, ad.ListResponse{
			ID:          adItem.ID,
			Title:       adItem.Title,
			Description: adItem.Description,
			ImageURL:    adItem.ImageURL,
			Price:       adItem.Price,
			AuthorLogin: adItem.AuthorLogin,
			IsOwner:     userID == adItem.AuthorID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
