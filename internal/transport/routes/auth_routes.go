package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/models"
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
		group.Get("/", r.Index())
		group.Get("/logout", r.Logout())
		group.Get("/login", r.LoginPage())
		group.Post("/login", r.ProcessLogin())

		// TODO: Feature flag
		group.Get("/register", r.Register())
		group.Post("/register", r.ProcessRegister())

		group.Get("/verify", r.VerifyEmail())

		// TODO: Feature flag
		group.Get("/forgot-password", r.ForgotPassword())
		group.Post("/forgot-password", r.ProcessForgotPassword())
	})

	router.Group(func(group chi.Router) {
		group.Use(middleware.NewAuthorizeMiddleware(r.sessionService))
		group.Get("/userinfo", r.UserInfo())
		group.Get("/home", r.Home())
		group.Post("/change-password", r.ChangePasswordLoggedIn())
		group.Post("/invite", r.Invite())
	})

	return "/auth", router
}

func (self *AuthRoutes) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "index.html", "layout", nil)
	}
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

		if user, err := self.authService.LoginUser(r.Context(), username, password); err == nil {
			id, err := self.sessionService.Create(r.Context(), user)
			if err != nil {
				panic(err)
			}

			// TODO: Add the cookie ttl to configuration
			http.SetCookie(w, utils.SessionCookie(id, 60))
			http.SetCookie(w, utils.ReturnToPostLoginCookie("", 0)) // Delete the cookie

			returnToCookie, err := r.Cookie(utils.RETURN_TO_COOKIE_NAME)
			if err != nil {
				http.Redirect(w, r, "/auth/home", http.StatusFound)
			} else {
				http.Redirect(w, r, returnToCookie.Value, http.StatusFound)
			}
		} else {
			// TODO: Flash to indicate failure
			http.Redirect(w, r, "/auth/login", http.StatusFound)
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

		self.sessionService.Destroy(r.Context(), cookie.Value)
		http.SetCookie(w, utils.SessionCookie("", 0))
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

func (self *AuthRoutes) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*models.User)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "home.html", "layout", user)
	}
}


func (self *AuthRoutes) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "register.html", "layout", nil)
	}
}

// TODO: Implement ME
func (self *AuthRoutes) ProcessRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// TODO: Implement ME
func (self *AuthRoutes) VerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// TODO: Implement ME
func (self *AuthRoutes) ChangePasswordLoggedIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (self *AuthRoutes) ForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "forgot_password.html", "layout", nil)
	}
}

// TODO: Implement ME
func (self *AuthRoutes) ProcessForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// TODO: Implement ME
func (self *AuthRoutes) Invite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

