package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type AuthRoutes struct {
	notificationConfig   config.NotificationsConfig
	sessionConfig        config.SessionConfig
	authService          services.AuthService
	sessionService       services.SessionService
	templateService      services.TemplateService
	accessControlService services.AccessControlService
}

func NewAuthRoutes(
	notificationConfig config.NotificationsConfig,
	sessionConfig config.SessionConfig,
	authService services.AuthService,
	sessionService services.SessionService,
	templateService services.TemplateService,
	accessControlService services.AccessControlService,
) *AuthRoutes {
	return &AuthRoutes{
		notificationConfig:   notificationConfig,
		sessionConfig:        sessionConfig,
		authService:          authService,
		sessionService:       sessionService,
		templateService:      templateService,
		accessControlService: accessControlService,
	}
}

func (r *AuthRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(r.sessionService))
	router.Get("/logout", r.Logout())
	router.Get("/password/change", r.ChangePassword())
	router.Post("/password/change", r.ProcessChangePassword())

	router.Group(func(group chi.Router) {
		group.Use(middleware.RedirectToHomeMiddleware)
		group.Get("/", r.Index())
		group.Get("/login", r.LoginPage())
		group.Post("/login", r.ProcessLogin())

		router.Get("/verify", r.VerifyEmail())
		router.Get("/verify/resend", r.ResendEmail())

		group.Get("/register", r.Register())
		group.Post("/register", r.ProcessRegister())

		group.Get("/password/forgot", r.ForgotPassword())
		group.Post("/password/forgot", r.ProcessForgotPassword())
	})

	router.Group(func(group chi.Router) {
		group.Use(middleware.RedirectToLoginMiddleware)
		group.Get("/userinfo", r.UserInfo())
		group.Get("/home", r.Home())
		group.Get("/invite", r.Invite())
		group.Post("/invite", r.ProcessInvite())
	})

	return "/auth", router
}

func (self *AuthRoutes) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/auth/home", http.StatusFound)
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
			self.templateService.Render(w, "userinfo.html", "layout", models.NewTemplateData(user))
		}
	}
}

func (self *AuthRoutes) LoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"login.html",
			"layout",
			models.NewTemplateError(utils.GetNotifications(r)),
		)
	}
}

func (self *AuthRoutes) ProcessLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("email")
		password := r.FormValue("password")

		user, err := self.authService.LoginUser(r.Context(), username, password)

		if err == nil {
			id := self.sessionService.Create(r.Context(), user)

			http.SetCookie(w, utils.SessionCookie(id, self.sessionConfig.CookieTTL))
			http.SetCookie(w, utils.ReturnToPostLoginCookie("", 0)) // Delete the cookie

			returnToCookie, err := r.Cookie(utils.RETURN_TO_COOKIE_NAME)
			if err != nil {
				http.Redirect(w, r, "/auth/home", http.StatusFound)
			} else {
				http.Redirect(w, r, returnToCookie.Value, http.StatusFound)
			}
		} else {
			utils.SetNotifications(w, err, "/auth/login", self.notificationConfig.Timeout)
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
		user := r.Context().Value("user").(*models.SessionData)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"home.html",
			"layout",
			models.NewTemplate(user, utils.GetNotifications(r)),
		)
	}
}

type RegisterData struct {
	Token string
	Id    string
}

func (self *AuthRoutes) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		id := r.URL.Query().Get("id")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"register.html",
			"layout",
			models.NewTemplate(RegisterData{token, id}, utils.GetNotifications(r)),
		)
	}
}

func (self *AuthRoutes) ProcessRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		token := r.FormValue("token")
		id := r.FormValue("id")
		url := fmt.Sprintf("/auth/register?token=%s&id=%s", token, id)

		err := self.authService.VerifyInvite(
			r.Context(),
			id,
			token,
			func(claims *models.InviteData) bool {
				return claims.Email == email
			},
		)
		if err != nil {
			utils.SetNotifications(w, err, "/auth/register", self.notificationConfig.Timeout)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		if password != confirmPassword {
			utils.SetNotifications(
				w,
				services.PasswordMismatch,
				"/auth/register",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		err = self.authService.CreateUser(r.Context(), username, email, password)
		if err != nil {
			utils.SetNotifications(w, err, "/auth/register", self.notificationConfig.Timeout)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		self.authService.InvalidateInvite(r.Context(), id)

		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

func (self *AuthRoutes) VerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		id := r.URL.Query().Get("id")

		if token == "" {
			utils.SetNotifications(
				w,
				services.InvalidRegistrationToken,
				"/auth/login",
				self.notificationConfig.Timeout,
			)
		} else {
			err := self.authService.VerifyUser(r.Context(), id, token)
			if err != nil {
				utils.SetNotifications(w, err, "/auth/login", self.notificationConfig.Timeout)
			}
		}

		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

func (self *AuthRoutes) ResendEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")

		if email == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			self.templateService.Render(
				w,
				"resend_email.html",
				"layout",
				models.NewTemplateError(utils.GetNotifications(r)),
			)
			return
		}

		err := self.authService.ResendVerifyEmail(r.Context(), email)
		if err != nil {
			utils.SetNotifications(w, err, "/auth/verify/resend", self.notificationConfig.Timeout)
			http.Redirect(w, r, "/auth/verify/resend", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

type tokenData struct {
	Token  string
	UserId string
}

type changePasswordAnonymousData struct {
	User      *models.SessionData
	TokenData *tokenData
}

func (self *AuthRoutes) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		var data changePasswordAnonymousData
		if user != nil {
			data = changePasswordAnonymousData{TokenData: nil, User: user.(*models.SessionData)}
		} else {
			token := r.URL.Query().Get("token")
			userId := r.URL.Query().Get("id")
			data = changePasswordAnonymousData{TokenData: &tokenData{token, userId}, User: nil}
		}
		self.templateService.Render(
			w,
			"change_password.html",
			"layout",
			models.NewTemplate(data, utils.GetNotifications(r)),
		)
	}
}

func (self *AuthRoutes) ProcessChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")

		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		token := r.FormValue("token")
		id := r.FormValue("id")
		var url string
		if user == nil {
			url = fmt.Sprintf("/auth/password/change?token=%s&id=%s", token, id)
		} else {
			url = "/auth/password/change"
		}

		if newPassword != confirmPassword {
			utils.SetNotifications(
				w,
				services.PasswordMismatch,
				"/auth/password/change",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		if user != nil {
			user := user.(*models.SessionData)
			currentPassword := r.FormValue("current_password")
			csrfToken := r.FormValue("csrf_token")

			if csrfToken != user.CsrfToken {
				utils.SetNotifications(
					w,
					utils.NewGenericMessage("Bad request, try again"),
					"/auth/password/change",
					self.notificationConfig.Timeout,
				)
				http.Redirect(w, r, url, http.StatusFound)
				return
			}

			err := self.authService.ChangePassword(
				r.Context(),
				user.UserId,
				currentPassword,
				newPassword,
			)
			if err != nil {
				utils.SetNotifications(
					w,
					err,
					"/auth/password/change",
					self.notificationConfig.Timeout,
				)
				http.Redirect(w, r, url, http.StatusFound)
				return
			}

			user.CsrfToken = uuid.New().String()
			self.sessionService.Update(r.Context(), user)
			if err != nil {
				panic(err)
			}
		} else {
			err := self.authService.ChangePasswordWithToken(
				r.Context(),
				id,
				token,
				newPassword,
			)
			if err != nil {
				utils.SetNotifications(w, err, "/auth/password/change", self.notificationConfig.Timeout)
				http.Redirect(w, r, url, http.StatusFound)
				return
			}
		}
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

func (self *AuthRoutes) ForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"forgot_password.html",
			"layout",
			models.NewTemplateError(utils.GetNotifications(r)),
		)
	}
}

func (self *AuthRoutes) ProcessForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		self.authService.CreateForgotPasswordToken(r.Context(), email)
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

func (self *AuthRoutes) Invite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*models.SessionData)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"invite_user.html",
			"layout",
			models.NewTemplate(
				map[string]interface{} { 
					"CsrfToken": user.CsrfToken, 
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *AuthRoutes) ProcessInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessControlErr := self.accessControlService.Enforce(r.Context(), "invite", "send")
		if accessControlErr != nil {
			utils.SetNotifications(
				w,
				accessControlErr,
				"/auth/invite",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/auth/invite", http.StatusFound)
			return
		}

		session := r.Context().Value("user").(*models.SessionData)
		email := r.FormValue("email")
		csrfToken := r.FormValue("csrf_token")

		if csrfToken != session.CsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("Bad request, try again"),
				"/auth/invite",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/auth/invite", http.StatusFound)
			return 
		}

		err := self.authService.InviteUser(r.Context(), session.UserId, email)

		if err != nil {
			utils.SetNotifications(w, err, "/auth/invite", self.notificationConfig.Timeout)
			http.Redirect(w, r, "/auth/invite", http.StatusFound)
			return
		}

		session.CsrfToken = uuid.New().String()
		self.sessionService.Update(r.Context(), session)

		http.Redirect(w, r, "/auth/home", http.StatusFound)
	}
}
