package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
)

type OauthRoutes struct {
	sessionService services.SessionService
	templateService services.TemplateService
}

func NewOauthRoutes(
	sessionService services.SessionService,
	templateService services.TemplateService,
) *OauthRoutes {
	return &OauthRoutes{
		sessionService: sessionService,
		templateService: templateService,
	}
}

// Routes implements transport.Router.
func (r *OauthRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(r.sessionService))

	// Application management
	router.Get("/applications/new", r.CreateApplication())
	router.Post("/applications", r.ProcessCreateApplication())
	router.Get("/applications/{id}", r.GetApplication())
	router.Get("/applications/{id}/update", r.UpdateApplication())
	router.Put("/applications/{id}", r.ProcessUpdateApplication())
	router.Delete("/applications/{id}", r.DeleteApplication())
	router.Get("/applications", r.ListApplications())

	// The actual Oauth Flow
	router.Get("/authorize", r.Authorize())
	router.Post("/token", r.Token())

	return "/oauth", router
}

func (self *OauthRoutes) CreateApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "application_create.html", "layout", models.NewTemplateEmpty())
	}
}

func (self *OauthRoutes) ProcessCreateApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (self *OauthRoutes) GetApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "application_detail.html", "layout", models.NewTemplateEmpty())
	}
}

func (self *OauthRoutes) UpdateApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "application_update.html", "layout", models.NewTemplateEmpty())
	}
}

func (self *OauthRoutes) ProcessUpdateApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (self *OauthRoutes) DeleteApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (self *OauthRoutes) ListApplications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "application_list.html", "layout", models.NewTemplateEmpty())
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
