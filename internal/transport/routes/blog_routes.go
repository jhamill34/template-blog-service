package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type BlogRoutes struct{}

func NewBlogRoutes() *BlogRoutes {
	return &BlogRoutes{}
}

func (b *BlogRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	return "/blog", router
}
