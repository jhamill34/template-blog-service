package utils

import (
	"net/http"
	"time"
)

const RETURN_TO_COOKIE_NAME = "return_to"

func ReturnToPostLoginCookie(location string, expires time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     RETURN_TO_COOKIE_NAME,
		Value:    location,
		Path:     "/auth/login",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(expires),
	}
}

const SESSION_COOKIE_NAME = "session_id"

func SessionCookie(id string, expires time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    id,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(expires),
	}
}
