package handler

import (
	"encoding/json"
	"net/http"
	"time"

	dto "github.com/AugustSerenity/marketplace/internal/handler/model"
	"github.com/AugustSerenity/marketplace/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	service Service
}

func New(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) Route() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /user-registration", h.UserRegistration)

	return router
}

// Принимаю регистрационные данные пользователя
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

	// Хэширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Сборка модели пользователя
	user := model.User{
		Login:        userData.Login,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	// Вызов сервиса
	err = h.service.RegistrationUser(r.Context(), &user)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// Возврат ответа
	resp := dto.RegistrationResponse{
		ID:    user.ID, // если ты получаешь ID от репозитория
		Login: user.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
