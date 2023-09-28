package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type AuthRoutes struct {
	authService     services.AuthService
	sessionService  services.SessionService
	templateService services.TemplateService
}

func NewAuthRoutes(
	authService services.AuthService,
	sessionService services.SessionService,
	templateService services.TemplateService,
) *AuthRoutes {
	return &AuthRoutes{
		authService:    authService,
		sessionService: sessionService,
		templateService: templateService,
	}
}

func (r *AuthRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	router.Group(func(group chi.Router) {
		group.Get("/logout", r.Logout())
		group.Get("/login", r.LoginPage())
		group.Post("/login", r.ProcessLogin())
	})

	router.Group(func(group chi.Router) {
		group.Use(middleware.NewAuthorizeMiddleware(r.sessionService))
		group.Get("/userinfo", r.UserInfo())
	})

	return "/auth", router
}

func (self *AuthRoutes) UserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")

		if r.Header.Get("Accept") == "application/json" {
			utils.RenderJSON(w, user, http.StatusOK)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			self.templateService.Render(w, "userinfo.html", "layout", user)
		}
	}
}

func (self *AuthRoutes) LoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "login.html", "layout", nil)
	}
}

func (self *AuthRoutes) ProcessLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("email")
		password := r.FormValue("password")

		if user, err := self.authService.LoginUser(username, password); err == nil {
			id := self.sessionService.Create(user)

			http.SetCookie(w, utils.SessionCookie(id, 60))
			http.SetCookie(w, utils.ReturnToPostLoginCookie("", 0)) // Delete the cookie

			returnToCookie, err := r.Cookie(utils.RETURN_TO_COOKIE_NAME)
			if err != nil {
				http.Redirect(w, r, "/auth/login", http.StatusFound)
			} else {
				http.Redirect(w, r, returnToCookie.Value, http.StatusFound)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func (self *AuthRoutes) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(utils.SESSION_COOKIE_NAME)
		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		self.sessionService.Destroy(cookie.Value)
		http.SetCookie(w, utils.SessionCookie("", 0))
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}
