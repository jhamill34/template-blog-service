package email

import (
	"context"
)

type AuthFromConfigHandler struct {
	Username string
	Password string
}

func NewAuthFromConfigHandler(username, password string) *AuthFromConfigHandler {
	return &AuthFromConfigHandler{username, password}
}

// Authenticate implements services.EmailAuthHandler.
func (self *AuthFromConfigHandler) Authenticate(
	ctx context.Context,
	username string,
	password string,
) bool {
	return username == self.Username && password == self.Password
}

