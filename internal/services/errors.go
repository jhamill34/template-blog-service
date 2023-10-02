package services

import (
	"github.com/jhamill34/notion-provisioner/internal/models"
)

type AuthServiceError struct {
	Message string
}

func (self *AuthServiceError) Notify() *models.Notification {
	return &models.Notification{Message: self.Message}
}

func NewAuthServiceError(message string) *AuthServiceError {
	return &AuthServiceError{Message: message}
}

var InvalidPassword *AuthServiceError = NewAuthServiceError("Invalid password")
var UnverifiedUser *AuthServiceError = NewAuthServiceError("User is not verified")
