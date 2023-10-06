package routes

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type OauthRoutes struct {
	appService         services.ApplicationService
	sessionService     services.SessionService
	templateService    services.TemplateService
	notificationConfig config.NotificationsConfig
}

func NewOauthRoutes(
	appService services.ApplicationService,
	sessionService services.SessionService,
	templateService services.TemplateService,
	notificationConfig config.NotificationsConfig,
) *OauthRoutes {
	return &OauthRoutes{
		appService:         appService,
		sessionService:     sessionService,
		templateService:    templateService,
		notificationConfig: notificationConfig,
	}
}

// Routes implements transport.Router.
func (r *OauthRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(r.sessionService))
	router.Post("/token", r.Token())
	router.Get("/verify", r.VerifyAccessToken())

	router.Group(func(group chi.Router) {
		group.Use(middleware.RedirectToLoginMiddleware)

		// Application management
		group.Get("/application/new", r.CreateApplication())
		group.Post("/application", r.ProcessCreateApplication())
		group.Get("/application/{id}", r.GetApplication())
		group.Delete("/application/{id}", r.DeleteApplication())
		group.Put("/application/{id}/secret", r.NewSecret())
		group.Get("/application", r.ListApplications())

		// The actual Oauth Flow
		group.Get("/authorize", r.Authorize())
	})

	return "/oauth", router
}

func (self *OauthRoutes) CreateApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("user").(*models.SessionData)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"application_create.html",
			"layout",
			models.NewTemplate(
				session,
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OauthRoutes) ProcessCreateApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("user").(*models.SessionData)

		name := r.FormValue("name")
		description := r.FormValue("description")
		redirectUri := r.FormValue("redirect_uri")
		csrfToken := r.FormValue("csrf_token")

		if csrfToken != session.CsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/oauth/application/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/oauth/application/new", http.StatusFound)
			return
		}

		_, clientSecret, err := self.appService.CreateApp(
			r.Context(),
			redirectUri,
			name,
			description,
		)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/oauth/application/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/oauth/application/new", http.StatusFound)
			return
		}

		session.CsrfToken = uuid.New().String()
		self.sessionService.Update(r.Context(), session)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"application_secret.html",
			"layout",
			models.NewTemplateData(clientSecret),
		)
	}
}

type GetAppData struct {
	CsrfToken string
	App       *models.App
}

func (self *OauthRoutes) GetApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("user").(*models.SessionData)
		id := chi.URLParam(r, "id")
		app, err := self.appService.GetApp(r.Context(), id)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/oauth/application",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/oauth/application", http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"application_detail.html",
			"layout",
			models.NewTemplate(GetAppData{session.CsrfToken, app}, utils.GetNotifications(r)),
		)
	}
}

func (self *OauthRoutes) DeleteApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("user").(*models.SessionData)
		id := chi.URLParam(r, "id")
		csrfToken := r.URL.Query().Get("csrf_token")

		if csrfToken != session.CsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/oauth/application",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/oauth/application")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := self.appService.DeleteApp(r.Context(), id); err != nil {
			utils.SetNotifications(
				w,
				err,
				"/oauth/application",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/oauth/application")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		session.CsrfToken = uuid.New().String()
		self.sessionService.Update(r.Context(), session)

		w.Header().Set("HX-Redirect", "/oauth/application")
		w.WriteHeader(http.StatusNoContent)
	}
}

type ApplicationListData struct {
	CsrfToken string
	Apps      []models.App
}

func (self *OauthRoutes) ListApplications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("user").(*models.SessionData)
		apps, err := self.appService.ListApps(r.Context())

		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/auth",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"application_list.html",
			"layout",
			models.NewTemplate(
				ApplicationListData{session.CsrfToken, apps},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OauthRoutes) NewSecret() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("user").(*models.SessionData)
		id := chi.URLParam(r, "id")
		csrfToken := r.URL.Query().Get("csrf_token")

		if csrfToken != session.CsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/oauth/application/"+id,
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/oauth/application/"+id)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		clientSecret, err := self.appService.NewSecret(r.Context(), id)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/oauth/application/"+id,
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/oauth/application/"+id)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"application_secret.html",
			"layout",
			models.NewTemplateData(clientSecret),
		)
	}
}

func (self *OauthRoutes) Authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response_type := r.URL.Query().Get("response_type")
		if response_type != "code" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		session := r.Context().Value("user").(*models.SessionData)

		client_id := r.URL.Query().Get("client_id")
		redirect_uri := r.URL.Query().Get("redirect_uri")
		state := r.URL.Query().Get("state")

		app, err := self.appService.GetAppByClientId(r.Context(), client_id)
		if err == services.AccessDenied {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err == services.AppNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if redirect_uri != app.RedirectUri {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		code := self.appService.NewAuthCode(r.Context(), session.UserId, app.AppId)

		http.Redirect(w, r, redirect_uri+"?code="+code+"&state="+state, http.StatusFound)
	}
}

func (self *OauthRoutes) Token() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		grantType := r.FormValue("grant_type")
		clientId := r.FormValue("client_id")
		clientSecret := r.FormValue("client_secret")

		switch grantType {
		case "authorization_code":
			code := r.FormValue("code")
			redirectUri := r.FormValue("redirect_uri")
			userId, appId, err := self.appService.GetAuthCode(r.Context(), code)
			if err != nil {
				log.Println("Invalid code: ", code)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			app, err := self.appService.ValidateAppSecret(r.Context(), appId, clientSecret)
			if err != nil {
				log.Println("Invalid client secret: ", clientSecret)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if app.ClientId != clientId {
				log.Println("Invalid client id: ", clientId)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if app.RedirectUri != redirectUri {
				log.Println("Invalid redirect uri: ", redirectUri)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			accessTokenResponse, err := self.appService.NewAccessToken(r.Context(), userId, appId, "")
			if err != nil {
				panic(err)
			}

			utils.RenderJSON(w, accessTokenResponse, http.StatusOK)
		case "refresh_token":
			refreshToken := r.FormValue("refresh_token")

			userId, appId, err := self.appService.FindRefreshToken(r.Context(), refreshToken)
			if err != nil {
				log.Println("Invalid refresh token: ", refreshToken)
				w.WriteHeader(http.StatusUnauthorized)
				return 
			}

			app, err := self.appService.ValidateAppSecret(r.Context(), appId, clientSecret)
			if err != nil {
				log.Println("Invalid client secret: ", clientSecret)
				w.WriteHeader(http.StatusUnauthorized)
				return 
			}

			if app.ClientId != clientId {
				log.Println("Invalid client id: ", clientId)
				w.WriteHeader(http.StatusUnauthorized)
				return 
			}

			accessTokenResponse, err := self.appService.NewAccessToken(r.Context(), userId, appId, refreshToken)
			if err != nil {
				panic(err)
			}

			utils.RenderJSON(w, accessTokenResponse, http.StatusOK)
		default:
			log.Println("Invalid grant type: ", grantType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func (self *OauthRoutes) VerifyAccessToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.URL.Query().Get("access_token")
		if accessToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !self.appService.VerifyAccessToken(r.Context(), accessToken) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// var _ transport.Router = (*OauthRoutes)(nil)
