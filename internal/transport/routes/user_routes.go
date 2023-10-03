package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type UserRoutes struct {
	sessionService  services.SessionService
	templateService services.TemplateService
	userService     services.UserService
}

func NewUserRoutes(
	sessionService services.SessionService,
	templateService services.TemplateService,
	userService services.UserService,
) *UserRoutes {
	return &UserRoutes{
		sessionService:  sessionService,
		templateService: templateService,
		userService:     userService,
	}
}

func (self *UserRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(self.sessionService))
	router.Use(middleware.RedirectToLoginMiddleware)

	router.Get("/", self.ListUsers())

	return "/user", router
}

func (self *UserRoutes) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := self.userService.ListUsers(r.Context())

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"users_list.html",
			"layout",
			models.NewTemplate(users, utils.GetNotifications(r)),
		)
	}
}
