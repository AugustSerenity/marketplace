package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	dto "github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	"github.com/AugustSerenity/marketplace/internal/middleware"
	"github.com/AugustSerenity/marketplace/internal/model"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
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

	var userData dto.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(userData); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Хэширование пароля
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

	// Регистрация пользователя
	err = h.service.RegistrationUser(r.Context(), &user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "User with this login already exists", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "reserved login") {
			http.Error(w, "Reserved login", http.StatusBadRequest)
			return
		}

		log.Printf("Registration DB error: %v", err)
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

	invalidTitleRegex := `[^a-zA-Z0-9\s]`
	if matched, _ := regexp.MatchString(invalidTitleRegex, req.Title); matched {
		http.Error(w, "Invalid title characters", http.StatusBadRequest)
		return
	}

	newAd := &model.Ad{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		AuthorID:    userID,
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

func (h *Handler) GetAds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	page := 1
	pageSize := 10

	if val := q.Get("page"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			page = parsed
		}
	}

	if val := q.Get("page_size"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			pageSize = parsed
		}
	}

	sortBy := q.Get("sort_by")
	sortOrder := q.Get("sort_order")

	minPrice := 0.0
	if val := q.Get("min_price"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			minPrice = parsed
		}
	}

	maxPrice := 0.0
	if val := q.Get("max_price"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			maxPrice = parsed
		}
	}

	if minPrice > maxPrice && maxPrice != 0 {
		http.Error(w, "min_price cannot be greater than max_price", http.StatusBadRequest)
		return
	}

	req := ad.ListRequest{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
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
