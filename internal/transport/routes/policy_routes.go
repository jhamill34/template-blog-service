package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/services/rbac"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type PolicyRoutes struct {
	sessionService       services.SessionService
	signer               services.Signer
	accessControlService services.AccessControlService
	policyProvider       rbac.PolicyProvider
}

func NewPolicyRoutes(
	sessionService services.SessionService,
	signer services.Signer,
	accessControlService services.AccessControlService,
	policyProvider rbac.PolicyProvider,
) *PolicyRoutes {
	return &PolicyRoutes{
		sessionService:       sessionService,
		signer:               signer,
		accessControlService: accessControlService,
		policyProvider:       policyProvider,
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

		if err := self.accessControlService.Enforce(r.Context(), "/user/"+userId+"/policy", "list"); err != nil {
			utils.RenderJSON(w, err, http.StatusForbidden)
			return
		}

		response, err := self.policyProvider.GetPolicies(r.Context(), userId)
		if err != nil {
			utils.RenderJSON(w, err, http.StatusInternalServerError)
			return
		}

		utils.RenderJSON(w, response, http.StatusOK)
	}
}

// var _ transport.Router = (*PolicyRoutes)(nil)
