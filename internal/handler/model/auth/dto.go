package auth

type RegistrationRequest struct {
	Login    string `json:"login" validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegistrationResponse struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}
