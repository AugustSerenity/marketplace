package auth

type LoginRequest struct {
	Login    string `json:"login" validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
