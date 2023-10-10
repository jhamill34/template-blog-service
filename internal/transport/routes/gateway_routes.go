package routes

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

type GatewayRoutes struct {
	sessionService services.SessionService
}

func NewGatewayRoutes(
	sessionService services.SessionService,
) *GatewayRoutes {
	return &GatewayRoutes{
		sessionService: sessionService,
	}
}

// Routes implements transport.Router.
func (self *GatewayRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	// TODO: Fix the session store to be more generic

	router.Get("/oauth/authorize", self.Authorize())
	router.Get("/oauth/callback", self.Callback())

	// TODO: Forward requests to app server and render response in templates

	return "/", router
}

func (self *GatewayRoutes) Authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Authorize")
	}
}

func (self *GatewayRoutes) Callback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Callback")
	}
}
