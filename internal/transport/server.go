package transport

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type Router interface {
	Routes() (string, http.Handler)
}

func mount(router *chi.Mux, routers ...Router) {
	for _, r := range routers {
		path, handler := r.Routes()
		router.Mount(path, handler)
	}
}

type Server struct {
	router *chi.Mux
	port   int
}

func NewServer(
	port int,
	routers ...Router,
) Server {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(httprate.LimitByIP(100, 1*time.Minute))

	mount(router, routers...)

	return Server{
		router: router,
		port:   port,
	}
}

func (s *Server) Start() {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	// TODO: Graceful shutdown
	log.Printf("Server listening on port %d\n", s.port)
	server.ListenAndServe()

	log.Print("\n")
}
