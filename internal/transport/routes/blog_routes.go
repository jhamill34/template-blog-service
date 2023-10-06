package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type BlogRoutes struct {
	postService        services.BlogPostService
	templateService    services.TemplateService
	notificationConfig config.NotificationsConfig
	signer             services.Signer
}

func NewBlogRoutes(
	postService services.BlogPostService,
	templateService services.TemplateService,
	notificationConfig config.NotificationsConfig,
	signer services.Signer,
) *BlogRoutes {
	return &BlogRoutes{
		postService:        postService,
		templateService:    templateService,
		notificationConfig: notificationConfig,
		signer:             signer,
	}
}

// Routes implements transport.Router.
func (self *BlogRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	router.Get("/", self.Index())
	router.Get("/blog", self.ListPosts())
	router.Get("/blog/{id}", self.GetPost())

	router.Group(func(group chi.Router) {
		group.Use(middleware.NewTokenAuthMiddleware(self.signer))
		group.Post("/blog", self.CreatePost())
		group.Put("/blog/{id}", self.UpdatePost())
		group.Delete("/blog/{id}", self.DeletePost())
	})

	return "/", router
}

func (self *BlogRoutes) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "index.html", "layout", models.NewTemplateEmpty())
	}
}

func (self *BlogRoutes) GetPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		post, err := self.postService.GetPost(r.Context(), id)
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

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "blog_detail.html", "layout", models.NewTemplateData(post))
	}
}

func (self *BlogRoutes) ListPosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts := self.postService.ListPosts(r.Context())

		w.WriteHeader(http.StatusOK)
		self.templateService.Render(w, "blog_list.html", "layout", models.NewTemplateData(posts))
	}
}

type CreatePostPayload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (self *BlogRoutes) CreatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*models.SessionData)

		var payload CreatePostPayload
		if r.Header.Get("Content-Type") == "application/json" {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			payload = CreatePostPayload{
				Title:   r.FormValue("title"),
				Content: r.FormValue("content"),
			}
		}

		post, err := self.postService.CreatePost(
			r.Context(),
			payload.Title,
			payload.Content,
			user.UserId,
		)
		if err != nil {
			panic(err)
		}

		utils.RenderJSON(w, post, http.StatusCreated)
	}
}

func (self *BlogRoutes) UpdatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (self *BlogRoutes) DeletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

// var _ transport.Router = (*BlogRoutes)(nil)
