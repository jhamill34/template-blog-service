package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

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

		log.Println(sessionData)

		ctx := r.Context()
		switch sessionData.Type {
		case "user":
			ctx = context.WithValue(ctx, "user_id", sessionData.Payload)
		case "token":
			var tokenData models.AccessTokenResponse
			json.Unmarshal([]byte(sessionData.Payload), &tokenData)
			ctx = context.WithValue(ctx, "token", &tokenData)
		}

		ctx = context.WithValue(ctx, "session_id", sessionId)
		ctx = context.WithValue(ctx, "csrf_token", sessionData.CsrfToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type TokenAuthMiddleware struct {
	signer services.Signer
}

func NewTokenAuthMiddleware(signer services.Signer) func(http.Handler) http.Handler {
	middleware := &TokenAuthMiddleware{signer}

	return middleware.AuthorizeMiddleware
}

func (m *TokenAuthMiddleware) AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, ok := r.Context().Value("user_id").(string)
		if ok && userId != "" {
			next.ServeHTTP(w, r)
			return 
		}

		token := r.Header.Get("Authorization")

		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if len(token) < 7 || token[:7] != "Bearer " {
			next.ServeHTTP(w, r)
			return
		}

		token = token[7:]

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			next.ServeHTTP(w, r)
			return
		}

		signature := parts[2]
		paylaod := parts[0] + "." + parts[1]

		if m.signer.Verify([]byte(paylaod), signature) != nil {
			next.ServeHTTP(w, r)
			return
		}

		claimsData, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		var claims models.AccessTokenClaims
		err = json.Unmarshal(claimsData, &claims)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		expiresAt := claims.Iat + claims.Exp
		if time.Now().Unix() > expiresAt {
			next.ServeHTTP(w, r)
			return
		}

		if claims.Iss != "auth" {
			next.ServeHTTP(w, r)
			return
		}


		ctx := context.WithValue(r.Context(), "user_id", claims.Sub)
		ctx = context.WithValue(ctx, "raw_token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UnauthorizedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user_id").(string)

		if !ok || user == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RedirectToIndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token")

		if token == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RedirectToLoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user_id").(string)

		if !ok || user == "" {
			http.SetCookie(w, utils.ReturnToPostLoginCookie(r.URL.String(), 5*time.Minute))
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RedirectToHomeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user_id").(string)

		if ok && user != "" {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
