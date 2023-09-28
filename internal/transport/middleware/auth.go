package middleware

import (
	"context"
	"net/http"

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
			http.SetCookie(w, utils.ReturnToPostLoginCookie(r.URL.Path, 5))
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		sessionId := cookie.Value
		sessionData, err := m.sessionService.Find(r.Context(), sessionId)
		if err != nil {
			http.SetCookie(w, utils.ReturnToPostLoginCookie(r.URL.Path, 5))
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "user", sessionData)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
