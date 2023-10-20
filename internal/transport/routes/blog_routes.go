package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/middleware"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type BlogRoutes struct {
	postService services.BlogPostService
	signer      services.Signer
}

func NewBlogRoutes(
	postService services.BlogPostService,
	signer services.Signer,
) *BlogRoutes {
	return &BlogRoutes{
		postService: postService,
		signer:      signer,
	}
}

// Routes implements transport.Router.
func (self *BlogRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	router.Get("/blog", self.ListPosts())
	router.Get("/blog/{id}", self.GetPost())

	router.Group(func(group chi.Router) {
		group.Use(middleware.NewTokenAuthMiddleware(self.signer))
		group.Use(middleware.UnauthorizedMiddleware)

		group.Post("/blog", self.CreatePost())
		group.Put("/blog/{id}", self.UpdatePost())
		group.Delete("/blog/{id}", self.DeletePost())
	})

	return "/", router
}

func (self *BlogRoutes) GetPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		post, err := self.postService.GetPost(r.Context(), id)
		if err == services.PostNotFound {
			utils.RenderJSON(
				w,
				models.ForwardError{Message: "Post Not Found"},
				http.StatusNotFound,
			)
			return
		}

		if err != nil {
			panic(err)
		}

		utils.RenderJSON(w, post, http.StatusOK)
	}
}

func (self *BlogRoutes) ListPosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts := self.postService.ListPosts(r.Context())

		utils.RenderJSON(w, posts, http.StatusOK)
	}
}

type PostPayload struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Image     string `json:"image"`
	ImageMIME string `json:"image_mime"`
}

func (self *BlogRoutes) CreatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("user_id").(string)

		var payload PostPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			utils.RenderJSON(
				w,
				models.ForwardError{Message: "Bad Request"},
				http.StatusBadRequest,
			)
			return
		}

		post, err := self.postService.CreatePost(
			r.Context(),
			payload.Title,
			payload.Content,
			userId,
		)
		if err == services.AccessDenied {
			utils.RenderJSON(
				w,
				models.ForwardError{Message: "Access denied"},
				http.StatusForbidden,
			)
			return
		}

		if err != nil {
			panic(err)
		}

		if payload.Image != "" {
			imageData, base64Err := base64.StdEncoding.DecodeString(payload.Image)
			if base64Err != nil {
				panic(err)
			}

			err = self.postService.AddImage(r.Context(), post.Id, payload.ImageMIME, imageData)
			if err == services.AccessDenied {
				utils.RenderJSON(
					w,
					models.ForwardError{Message: "Access denied"},
					http.StatusForbidden,
				)
				return
			}
			if err != nil {
				panic(err)
			}
		}

		utils.RenderJSON(w, post, http.StatusCreated)
	}
}

func (self *BlogRoutes) UpdatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var payload PostPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			utils.RenderJSON(
				w,
				models.ForwardError{Message: "Bad Request"},
				http.StatusBadRequest,
			)
			return
		}

		post, err := self.postService.UpdatePost(
			r.Context(),
			id,
			payload.Title,
			payload.Content,
		)
		if err == services.AccessDenied {
			utils.RenderJSON(
				w,
				models.ForwardError{Message: "Access denied"},
				http.StatusForbidden,
			)
			return
		}

		if err != nil {
			panic(err)
		}
		
		if payload.Image != "" {
			imageData, base64Err := base64.StdEncoding.DecodeString(payload.Image)
			if base64Err != nil {
				panic(err)
			}

			err = self.postService.AddImage(r.Context(), post.Id, payload.ImageMIME, imageData)
			if err == services.AccessDenied {
				utils.RenderJSON(
					w,
					models.ForwardError{Message: "Access denied"},
					http.StatusForbidden,
				)
				return
			}
			if err != nil {
				panic(err)
			}
		}

		utils.RenderJSON(w, post, http.StatusCreated)
	}
}

func (self *BlogRoutes) DeletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := self.postService.DeletePost(r.Context(), id)
		if err == services.AccessDenied {
			utils.RenderJSON(
				w,
				models.ForwardError{Message: "Access denied"},
				http.StatusForbidden,
			)
			return
		}

		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
