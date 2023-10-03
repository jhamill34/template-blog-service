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
var InvalidInviteToken *AuthServiceError = NewAuthServiceError("Invalid invite token")
var InvalidPasswordToken *AuthServiceError = NewAuthServiceError("Invalid password token")
var InvalidRegistrationToken *AuthServiceError = NewAuthServiceError("Invalid registration token")
var PasswordMismatch *AuthServiceError = NewAuthServiceError("Password mismatch")
var EmailAlreadyInUse *AuthServiceError = NewAuthServiceError("Email already in use")
var AccountAlreadyVerified *AuthServiceError = NewAuthServiceError("Account already verified")
var AccountNotFound *AuthServiceError = NewAuthServiceError("Account not found")


//================================================== 

type TokenError struct {
	Message string
}

func (self *TokenError) Notify() *models.Notification {
	return &models.Notification{Message: self.Message}
}

func NewTokenError(message string) *TokenError {
	return &TokenError{Message: message}
}

var InvalidToken *TokenError = NewTokenError("Invalid token")
var TokenNotFound *TokenError = NewTokenError("Token not found")

