package routes

import (
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
	router.Use(middleware.RedirectToLoginMiddleware)

	// Application management
	router.Get("/application/new", r.CreateApplication())
	router.Post("/application", r.ProcessCreateApplication())
	router.Get("/application/{id}", r.GetApplication())
	router.Delete("/application/{id}", r.DeleteApplication())
	router.Get("/application", r.ListApplications())

	// The actual Oauth Flow
	router.Get("/authorize", r.Authorize())
	router.Post("/token", r.Token())

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

func (self *OauthRoutes) GetApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			models.NewTemplate(app, utils.GetNotifications(r)),
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

func (self *OauthRoutes) Authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (self *OauthRoutes) Token() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// var _ transport.Router = (*OauthRoutes)(nil)
