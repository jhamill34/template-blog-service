package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type PolicyRoutes struct {
	sessionService services.SessionService
	signer         services.Signer
	userService    services.UserService
}

func NewPolicyRoutes(
	sessionService services.SessionService,
	signer services.Signer,
	userService services.UserService,
) *PolicyRoutes {
	return &PolicyRoutes{
		sessionService: sessionService,
		signer:         signer,
		userService:    userService,
	}
}

// Routes implements transport.Router.
func (self *PolicyRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(self.sessionService))
	router.Use(middleware.NewTokenAuthMiddleware(self.signer))
	router.Use(middleware.UnauthorizedMiddleware)

	router.Get("/", self.ListMyPolicies())
	return "/policy", router
}

func (self *PolicyRoutes) ListMyPolicies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("user_id").(string)

		policies, err := self.userService.ListPolicies(r.Context(), userId)
		if err != nil {
			panic(err)
		}

		response := models.PolicyResponse{
			User: policies,
		}

		utils.RenderJSON(w, response, http.StatusOK)
	}
}

// var _ transport.Router = (*PolicyRoutes)(nil)
