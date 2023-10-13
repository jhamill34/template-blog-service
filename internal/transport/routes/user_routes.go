package routes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type UserRoutes struct {
	notificationConfig config.NotificationsConfig
	sessionService     services.SessionService
	templateService    services.TemplateService
	userService        services.UserService
}

func NewUserRoutes(
	notificationConfig config.NotificationsConfig,
	sessionService services.SessionService,
	templateService services.TemplateService,
	userService services.UserService,
) *UserRoutes {
	return &UserRoutes{
		notificationConfig: notificationConfig,
		sessionService:     sessionService,
		templateService:    templateService,
		userService:        userService,
	}
}

func (self *UserRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(self.sessionService))
	router.Use(middleware.RedirectToLoginMiddleware)

	router.Get("/", self.ListUsers())
	router.Get("/{id}", self.GetUser())
	router.Get("/{id}/policy", self.ListPolicies())
	router.Post("/{id}/policy", self.ProcessCreatePolicy())
	router.Get("/{id}/policy/new", self.CreatePolicy())
	router.Delete("/{id}/policy/{policyId}", self.DeletePolicy())

	return "/user", router
}

func (self *UserRoutes) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := self.userService.ListUsers(r.Context())
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

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"users_list.html",
			"layout",
			models.NewTemplate(users, utils.GetNotifications(r)),
		)
	}
}

func (self *UserRoutes) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "id")

		user, err := self.userService.GetUser(r.Context(), userId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/user",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/user", http.StatusFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"users_view.html",
			"layout",
			models.NewTemplate(user, utils.GetNotifications(r)),
		)
	}
}

type PolicyListData struct {
	UserId    string
	CsrfToken string
	Policies  []models.Policy
}

func (self *UserRoutes) ListPolicies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		userId := chi.URLParam(r, "id")

		policies, err := self.userService.ListPolicies(r.Context(), userId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/user/"+userId,
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/user/"+userId, http.StatusFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"users_policy_list.html",
			"layout",
			models.NewTemplate(
				PolicyListData{userId, userCsrfToken, policies},
				utils.GetNotifications(r),
			),
		)
	}
}

type NewPolicyData struct {
	UserId    string
	CsrfToken string
}

func (self *UserRoutes) CreatePolicy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		userId := chi.URLParam(r, "id")

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"users_policy_create.html",
			"layout",
			models.NewTemplate(NewPolicyData{userId, userCsrfToken}, utils.GetNotifications(r)),
		)
	}
}

func (self *UserRoutes) ProcessCreatePolicy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)
		userId := chi.URLParam(r, "id")

		resource := r.FormValue("resource")
		action := r.FormValue("action")
		effect := r.FormValue("effect")
		csrfToken := r.FormValue("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("Bad request, try again"),
				"/user/"+userId+"/policy/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/user/"+userId+"/policy/new", http.StatusSeeOther)
			return
		}

		if err := self.userService.CreatePolicy(r.Context(), userId, resource, action, effect); err != nil {
			utils.SetNotifications(
				w,
				err,
				"/user/"+userId+"/policy/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/user/"+userId+"/policy/new", http.StatusSeeOther)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		http.Redirect(w, r, "/user/"+userId+"/policy", http.StatusSeeOther)
	}
}

func (self *UserRoutes) DeletePolicy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)
		userId := chi.URLParam(r, "id")
		policyId, err := strconv.ParseInt(chi.URLParam(r, "policyId"), 10, 64)
		if err != nil {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("Bad request, try again"),
				"/user/"+userId+"/policy",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/user/"+userId+"/policy")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		csrfToken := r.URL.Query().Get("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("Bad request, try again"),
				"/user/"+userId+"/policy",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/user/"+userId+"/policy")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := self.userService.DeletePolicy(r.Context(), userId, int(policyId)); err != nil {
			utils.SetNotifications(
				w,
				err,
				"/user/"+userId+"/policy",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/user/"+userId+"/policy")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		w.Header().Set("HX-Redirect", "/user/"+userId+"/policy")
		w.WriteHeader(http.StatusNoContent)
	}
}
