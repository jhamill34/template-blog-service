package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type GatewayRoutes struct {
	sessionService     services.SessionService
	templateService    services.TemplateService
	httpClient         *http.Client
	sessionConfig      config.SessionConfig
	oauthConfig        config.OauthConfig
	authServer         string
	externalAuthServer string
	appServer          string
	notificationConfig config.NotificationsConfig
}

func NewGatewayRoutes(
	sessionService services.SessionService,
	templateService services.TemplateService,
	httpClient *http.Client,
	sessionConfig config.SessionConfig,
	oauthConfig config.OauthConfig,
	authServer string,
	externalAuthServer string,
	appServer string,
	notificationConfig config.NotificationsConfig,
) *GatewayRoutes {
	return &GatewayRoutes{
		sessionService:     sessionService,
		templateService:    templateService,
		httpClient:         httpClient,
		sessionConfig:      sessionConfig,
		oauthConfig:        oauthConfig,
		authServer:         authServer,
		externalAuthServer: externalAuthServer,
		appServer:          appServer,
		notificationConfig: notificationConfig,
	}
}

// Routes implements transport.Router.
func (self *GatewayRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	router.Get("/", self.Index())
	router.Get("/oauth/authorize", self.Authorize())
	router.Get("/oauth/callback", self.Callback())
	router.Get("/logout", self.Logout())

	router.Get("/blog", self.ListPosts())
	router.Get("/blog/{id}", self.GetPost())

	router.Group(func(group chi.Router) {
		group.Use(middleware.NewAuthorizeMiddleware(self.sessionService))
		group.Use(middleware.RedirectToIndexMiddleware)
		group.Get("/blog/new", self.NewPost())
		group.Post("/blog", self.ProcessNewPost())
	})

	return "/", router
}

func (self *GatewayRoutes) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/blog", http.StatusFound)
	}
}

func (self *GatewayRoutes) ListPosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var posts []models.PostStub
		var err *models.Notification
		err = self.forward(r, &posts)
		if err == nil {
			err = utils.GetNotifications(r)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"blog_list.html",
			"layout",
			models.NewTemplate(&posts, err),
		)
	}
}

func (self *GatewayRoutes) GetPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var post models.PostContent
		var err *models.Notification
		err = self.forward(r, &post)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/blog",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/blog", http.StatusFound)
			return 
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"blog_detail.html",
			"layout",
			models.NewTemplate(&post, utils.GetNotifications(r)),
		)
	}
}

func (self *GatewayRoutes) NewPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := r.Context().Value("csrf_token").(string)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(
			w,
			"blog_new.html",
			"layout",
			models.NewTemplate(
				map[string]string{
					"CsrfToken": csrfToken,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *GatewayRoutes) ProcessNewPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := self.forward(r, nil)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/blog/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/blog/new", http.StatusFound)
			return
		}
		
		sessionId := r.Context().Value("session_id").(string)
		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		http.Redirect(w, r, "/blog", http.StatusFound)
	}
}

type ForwardError struct {
	Message string `json:"message"`
}

// Notify implements models.Notifier.
func (self *ForwardError) Notify() *models.Notification {
	return &models.Notification{
		Message: self.Message,
	}
}

func (self *GatewayRoutes) forward(r *http.Request, data interface{}) *models.Notification {
	endpoint, err := url.Parse(r.URL.String())
	if err != nil {
		panic(err)
	}
	if r.URL.Scheme == "" {
		endpoint.Scheme = "http"
	} else {
		endpoint.Scheme = r.URL.Scheme
	}
	endpoint.Host = self.appServer

	appReq, err := http.NewRequestWithContext(r.Context(), r.Method, endpoint.String(), r.Body)
	if err != nil {
		panic(err)
	}
	appReq.Header = r.Header

	tokenData, ok := r.Context().Value("token").(*models.AccessTokenResponse)
	if ok && tokenData != nil {
		appReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))
	}

	appres, err := self.httpClient.Do(appReq)
	if err != nil {
		panic(err)
	}

	if  appres.StatusCode >= 200 && appres.StatusCode < 300 {
		if data != nil {
			json.NewDecoder(appres.Body).Decode(data)
		}
		return nil
	}

	var forwardError models.Notification 
	json.NewDecoder(appres.Body).Decode(&forwardError)
	return &forwardError
}

func (self *GatewayRoutes) Authorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoint, err := url.Parse(self.externalAuthServer + self.oauthConfig.AuthorizeUri)
		if err != nil {
			panic(err)
		}

		data := url.Values{}
		data.Set("response_type", "code")
		data.Set("client_id", self.oauthConfig.ClientID)
		data.Set("redirect_uri", "http://localhost:3332/oauth/callback")
		data.Set("state", "testing")
		endpoint.RawQuery = data.Encode()

		http.Redirect(w, r, endpoint.String(), http.StatusTemporaryRedirect)
	}
}

func (self *GatewayRoutes) Callback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		tokenEndpoint := fmt.Sprintf("%s%s", self.authServer, self.oauthConfig.TokenUri)

		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", "http://localhost:3332/oauth/callback")
		data.Set("client_id", self.oauthConfig.ClientID)
		data.Set("client_secret", self.oauthConfig.ClientSecret)

		tokenReq, err := http.NewRequestWithContext(
			r.Context(),
			"POST",
			tokenEndpoint,
			bytes.NewBufferString(data.Encode()),
		)
		if err != nil {
			panic(err)
		}

		tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		tokenResponse, err := self.httpClient.Do(tokenReq)
		if err != nil {
			panic(err)
		}

		if tokenResponse.StatusCode != http.StatusOK {
			panic(fmt.Errorf("Unexpected status code: %d", tokenResponse.StatusCode))
		}

		body, err := io.ReadAll(tokenResponse.Body)
		if err != nil {
			panic(err)
		}

		sessionId := self.sessionService.Create(r.Context(), &models.SessionData{
			Payload:   string(body),
			Type:      "token",
			CsrfToken: uuid.New().String(),
		})

		http.SetCookie(w, utils.SessionCookie(sessionId, self.sessionConfig.CookieTTL))
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (self *GatewayRoutes) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(utils.SESSION_COOKIE_NAME)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		self.sessionService.Destroy(r.Context(), cookie.Value)
		http.SetCookie(w, utils.SessionCookie("", 0))
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
