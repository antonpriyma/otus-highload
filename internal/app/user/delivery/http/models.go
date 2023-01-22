package http

type UserLoginRequest struct {
	ID       string `json:"id" validate:"required"`
	Password string `json:"password" validate:"required"`
}
