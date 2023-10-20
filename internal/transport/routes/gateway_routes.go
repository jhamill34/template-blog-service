package routes

import (
	"bytes"
	"encoding/base64"
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
	baseUrl            string
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
	baseUrl string,
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
		baseUrl:            baseUrl,
	}
}

// Routes implements transport.Router.
func (self *GatewayRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()
	router.Use(middleware.NewAuthorizeMiddleware(self.sessionService))

	router.Get("/", self.Index())
	router.Get("/oauth/authorize", self.Authorize())
	router.Get("/oauth/callback", self.Callback())
	router.Get("/logout", self.Logout())

	router.Get("/blog", self.ListPosts())
	router.Get("/blog/{id}", self.GetPost())

	router.Group(func(group chi.Router) {
		group.Use(middleware.RedirectToIndexMiddleware)
		group.Get("/blog/new", self.NewPost())
		group.Post("/blog", self.ProcessNewPost())

		group.Get("/blog/{id}/edit", self.EditPost())
		group.Put("/blog/{id}", self.ProcessEditPost())

		group.Delete("/blog/{id}", self.ProcessDeletePost())
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

		var response bytes.Buffer
		err = self.forward(r, nil, nil, &response)
		json.NewDecoder(&response).Decode(&posts)

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

		var response bytes.Buffer
		err = self.forward(r, nil, nil, &response)
		json.NewDecoder(&response).Decode(&post)

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

func (self *GatewayRoutes) EditPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := r.Context().Value("csrf_token").(string)
		id := chi.URLParam(r, "id")
		endpoint := "/blog/" + id

		var post models.PostContent

		var response bytes.Buffer
		err := self.forward(r, &endpoint, nil, &response)
		json.NewDecoder(&response).Decode(&post)

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
			"blog_edit.html",
			"layout",
			models.NewTemplate(
				map[string]interface{}{
					"CsrfToken": csrfToken,
					"Post":      post,
				},
				utils.GetNotifications(r),
			),
		)
	}
}

func (self *GatewayRoutes) ProcessEditPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userCsrfToken := r.Context().Value("csrf_token").(string)
		csrfToken := r.FormValue("csrf_token")
		if userCsrfToken != csrfToken {
			utils.SetNotifications(
				w,
				&models.Notification{
					Message: "Bad Request",
				},
				"/blog/"+id,
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/blog/"+id)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var payload bytes.Buffer
		jsonValue := FromUrlValues(r.Form)
		jsonValue.Encode(&payload)
		err := self.forward(r, nil, &payload, nil)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/blog/"+id,
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/blog/"+id)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		sessionId := r.Context().Value("session_id").(string)
		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		w.Header().Set("HX-Redirect", "/blog/"+id)
		w.WriteHeader(http.StatusNoContent)
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
		userCsrfToken := r.Context().Value("csrf_token").(string)
		csrfToken := r.FormValue("csrf_token")
		if userCsrfToken != csrfToken {
			utils.SetNotifications(
				w,
				&models.Notification{
					Message: "Bad Request",
				},
				"/blog/new",
				self.notificationConfig.Timeout,
			)
			http.Redirect(w, r, "/blog/new", http.StatusFound)
			return
		}

		var payload bytes.Buffer
		jsonValue := FromUrlValues(r.Form)

		file, header, fileErr := r.FormFile("image")
		if fileErr == nil {
			defer file.Close()

			var image bytes.Buffer
			io.Copy(&image, file)
			jsonValue["image"] = base64.StdEncoding.EncodeToString(image.Bytes())
			jsonValue["image_mime"] = header.Header.Get("Content-Type")
		}

		jsonValue.Encode(&payload)

		err := self.forward(r, nil, &payload, nil)
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

func (self *GatewayRoutes) ProcessDeletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userCsrfToken := r.Context().Value("csrf_token").(string)
		csrfToken := r.URL.Query().Get("csrf_token")
		if userCsrfToken != csrfToken {
			utils.SetNotifications(
				w,
				&models.Notification{
					Message: "Bad Request",
				},
				"/blog/"+id+"/edit",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/blog/"+id+"/edit")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		err := self.forward(r, nil, nil, nil)
		if err != nil {
			utils.SetNotifications(
				w,
				err,
				"/blog/"+id+"/edit",
				self.notificationConfig.Timeout,
			)
			w.Header().Set("HX-Redirect", "/blog/"+id+"/edit")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		sessionId := r.Context().Value("session_id").(string)
		self.sessionService.UpdateCsrf(r.Context(), sessionId, uuid.New().String())

		w.Header().Set("HX-Redirect", "/blog")
		w.WriteHeader(http.StatusNoContent)
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

type JSONValue map[string]interface{}

func FromUrlValues(values url.Values) JSONValue {
	jsonValue := make(JSONValue)

	for key, value := range values {
		if len(value) == 1 {
			jsonValue[key] = value[0]
		} else {
			jsonValue[key] = value
		}
	}
	return jsonValue
}

func (self *JSONValue) Encode(w io.Writer) error {
	return json.NewEncoder(w).Encode(self)
}

func (self *GatewayRoutes) forward(
	r *http.Request,
	toUrl *string,
	reqData io.Reader,
	resData io.Writer,
) *models.Notification {
	var endpointString string
	if toUrl != nil {
		endpointString = *toUrl
	} else {
		endpointString = r.URL.String()
	}
	endpoint, err := url.Parse(endpointString)
	if err != nil {
		panic(err)
	}
	if r.URL.Scheme == "" {
		endpoint.Scheme = "http"
	} else {
		endpoint.Scheme = r.URL.Scheme
	}
	endpoint.Host = self.appServer

	appReq, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		endpoint.String(),
		reqData,
	)
	if err != nil {
		panic(err)
	}
	appReq.Header = r.Header

	appReq.Header.Set("Content-Type", "application/json")
	tokenData, ok := r.Context().Value("token").(*models.AccessTokenResponse)
	if ok && tokenData != nil {
		appReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))
	}

	appres, err := self.httpClient.Do(appReq)
	if err != nil {
		panic(err)
	}

	if appres.StatusCode >= 200 && appres.StatusCode < 300 {
		if resData != nil {
			if _, err := io.Copy(resData, appres.Body); err != nil {
				panic(err)
			}
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
		data.Set("redirect_uri", self.baseUrl+"/oauth/callback")
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
		data.Set("redirect_uri", self.baseUrl+"/oauth/callback")
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
