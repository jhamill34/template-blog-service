package middleware

import (
	"context"
	"net/http"

	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type AuthorizeMiddleware struct {
	sessionService services.SessionService
}

func NewAuthorizeMiddleware(
	sessionService services.SessionService,
) func(http.Handler) http.Handler {
	middleware := &AuthorizeMiddleware{
		sessionService: sessionService,
	}

	return middleware.AuthorizeMiddleware
}

func (m *AuthorizeMiddleware) AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(utils.SESSION_COOKIE_NAME)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		var sessionData models.SessionData
		sessionId := cookie.Value
		sessionErr := m.sessionService.Find(r.Context(), sessionId, &sessionData)
		if sessionErr != nil {
			http.SetCookie(w, utils.SessionCookie("", 0))
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "user", &sessionData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RedirectToLoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")

		if user == nil {
			http.SetCookie(w, utils.ReturnToPostLoginCookie(r.URL.Path, 5))
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RedirectToHomeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")

		if user != nil {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
