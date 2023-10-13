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

type OrganizationRoutes struct {
	notificationConfig config.NotificationsConfig
	sessionService     services.SessionService
	templateService    services.TemplateService
	orgService         services.OrganizationService
}

func NewOrganizationRoutes(
	notificationConfig config.NotificationsConfig,
	sessionService services.SessionService,
	templateService services.TemplateService,
	orgService services.OrganizationService,
) *OrganizationRoutes {
	return &OrganizationRoutes{
		notificationConfig: notificationConfig,
		sessionService:     sessionService,
		templateService:    templateService,
		orgService:         orgService,
	}
}

// Routes implements transport.Router.
func (self *OrganizationRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(self.sessionService))
	router.Use(middleware.RedirectToLoginMiddleware)

	router.Get("/", self.ListMyOrgs())

	router.Get("/new", self.CreateOrg())
	router.Post("/", self.ProcessCreateOrg())

	router.Get("/{id}", self.GetOrg())
	router.Delete("/{id}", self.DeleteOrg())

	router.Get("/{id}/policy", self.ListOrgPolicies())

	router.Get("/{id}/policy/new", self.CreateOrgPolicy())
	router.Post("/{id}/policy", self.ProcessCreateOrgPolicy())
	router.Delete("/{id}/policy/{policyId}", self.DeleteOrgPolicy())

	router.Get("/{id}/user", self.ListOrgUsers())
	router.Get("/{id}/user/new", self.InviteUserToOrg())
	router.Post("/{id}/user", self.ProcessInviteUserToOrg())
	router.Get("/join", self.JoinOrg())
	router.Delete("/{id}/user/{userId}", self.RemoveUserFromOrg())

	return "/org", router
}

type ListOrgsData struct {
	CsrfToken string
	Orgs      []models.Organization
}

func (self *OrganizationRoutes) ListMyOrgs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		userId := r.Context().Value("user_id").(string)

		orgs, err := self.orgService.ListUsersOrgs(r.Context(), userId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/auth",
				self.notificationConfig.Timeout,
			)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"org_list.html",
			"layout",
			models.NewTemplate(
				ListOrgsData{
					CsrfToken: userCsrfToken,
					Orgs:      orgs,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OrganizationRoutes) CreateOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"org_create.html",
			"layout",
			models.NewTemplate(
				map[string]string{
					"CsrfToken": userCsrfToken,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OrganizationRoutes) ProcessCreateOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("user_id").(string)
		sessionId := r.Context().Value("session_id").(string)
		userCsrfToken := r.Context().Value("csrf_token").(string)

		name := r.FormValue("name")
		description := r.FormValue("description")
		csrfToken := r.FormValue("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/oauth/application/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/oauth/application/new", http.StatusFound)
			return
		}

		err := self.orgService.CreateOrganization(r.Context(), userId, name, description)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/new", http.StatusFound)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())
		http.Redirect(w, r, "/org", http.StatusFound)
	}
}

type GetOrgData struct {
	CsrfToken string
	Org       *models.Organization
}

func (self *OrganizationRoutes) GetOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		orgId := chi.URLParam(r, "id")

		org, err := self.orgService.GetOrganizationBydId(r.Context(), orgId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org", http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"org_detail.html",
			"layout",
			models.NewTemplate(
				GetOrgData{
					CsrfToken: userCsrfToken,
					Org:       org,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OrganizationRoutes) DeleteOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)

		orgId := chi.URLParam(r, "id")
		csrfToken := r.URL.Query().Get("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/org",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId, http.StatusFound)
			return
		}

		err := self.orgService.DeleteOrganization(r.Context(), orgId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org", http.StatusFound)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		w.Header().Set("HX-Redirect", "/org")
		w.WriteHeader(http.StatusNoContent)
	}
}

type ListOrgPoliciesData struct {
	CsrfToken string
	Policies  []models.Policy
	OrgId     string
}

func (self *OrganizationRoutes) ListOrgPolicies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		orgId := chi.URLParam(r, "id")

		policies, err := self.orgService.ListPolicies(r.Context(), orgId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/"+orgId,
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId, http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"org_policy_list.html",
			"layout",
			models.NewTemplate(
				ListOrgPoliciesData{
					CsrfToken: userCsrfToken,
					Policies:  policies,
					OrgId:     orgId,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

type CreateInOrgData struct {
	CsrfToken string
	OrgId     string
}

func (self *OrganizationRoutes) CreateOrgPolicy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		orgId := chi.URLParam(r, "id")

		self.templateService.Render(
			w,
			"org_policy_create.html",
			"layout",
			models.NewTemplate(
				CreateInOrgData{
					CsrfToken: userCsrfToken,
					OrgId:     orgId,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OrganizationRoutes) ProcessCreateOrgPolicy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)

		resource := r.FormValue("resource")
		action := r.FormValue("action")
		effect := r.FormValue("effect")
		orgId := chi.URLParam(r, "id")

		csrfToken := r.FormValue("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/org/"+orgId+"/policy/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId+"/policy/new", http.StatusFound)
			return
		}

		err := self.orgService.CreatePolicy(r.Context(), orgId, resource, action, effect)

		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/"+orgId+"/policy/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId+"/policy/new", http.StatusFound)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())
		http.Redirect(w, r, "/org/"+orgId+"/policy", http.StatusFound)
	}
}

func (self *OrganizationRoutes) DeleteOrgPolicy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)

		id := chi.URLParam(r, "id")
		policyId, parseIntErr := strconv.ParseInt(chi.URLParam(r, "policyId"), 10, 64)
		if parseIntErr != nil {
			panic(parseIntErr)
		}

		csrfToken := r.URL.Query().Get("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/org/"+id+"/policy",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+id+"/policy", http.StatusFound)
			return
		}

		err := self.orgService.DeletePolicy(r.Context(), id, int(policyId))
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/"+id+"/policy",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+id+"/policy", http.StatusFound)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		w.Header().Set("HX-Redirect", "/org/"+id+"/policy")
		w.WriteHeader(http.StatusNoContent)
	}
}

type ListOrgUsersData struct {
	CsrfToken string
	Users     []models.User
	OrgId     string
}

func (self *OrganizationRoutes) ListOrgUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		orgId := chi.URLParam(r, "id")

		users, err := self.orgService.ListUsers(r.Context(), orgId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/"+orgId,
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId, http.StatusFound)
			return 
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"org_users_list.html",
			"layout",
			models.NewTemplate(
				ListOrgUsersData{
					CsrfToken: userCsrfToken,
					OrgId:     orgId,
					Users:     users,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OrganizationRoutes) InviteUserToOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		orgId := chi.URLParam(r, "id")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"org_users_create.html",
			"layout",
			models.NewTemplate(
				CreateInOrgData{
					CsrfToken: userCsrfToken,
					OrgId:     orgId,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *OrganizationRoutes) ProcessInviteUserToOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)

		orgId := chi.URLParam(r, "id")
		email := r.FormValue("email")
		csrfToken := r.FormValue("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/org/"+orgId+"/user/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId+"/user/new", http.StatusFound)
			return
		}

		err := self.orgService.InviteUser(r.Context(), orgId, email)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/"+orgId+"/user/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId+"/user/new", http.StatusFound)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())
		http.Redirect(w, r, "/org/"+orgId+"/user", http.StatusFound)
	}
}

func (self *OrganizationRoutes) JoinOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("user_id").(string)

		token := r.URL.Query().Get("token")
		tokenId := r.URL.Query().Get("token_id")


		err := self.orgService.Join(r.Context(), tokenId, token, userId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/org", http.StatusFound)
	}
}

func (self *OrganizationRoutes) RemoveUserFromOrg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCsrfToken := r.Context().Value("csrf_token").(string)
		sessionId := r.Context().Value("session_id").(string)

		orgId := chi.URLParam(r, "id")
		userId := chi.URLParam(r, "userId")
		csrfToken := r.URL.Query().Get("csrf_token")

		if csrfToken != userCsrfToken {
			utils.SetNotifications(
				w,
				utils.NewGenericMessage("bad request, please try again."),
				"/org/"+orgId+"/user",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId+"/user", http.StatusFound)
			return
		}

		err := self.orgService.RemoveUser(r.Context(), orgId, userId)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/org/"+orgId+"/user",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/org/"+orgId+"/user", http.StatusFound)
			return
		}

		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		w.Header().Set("HX-Redirect", "/org/"+orgId+"/user")
		w.WriteHeader(http.StatusNoContent)
	}
}
