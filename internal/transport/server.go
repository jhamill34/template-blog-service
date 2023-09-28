package transport

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func (s *Server) Start(ctx context.Context) {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	shutdownSignal := handleShutdown(func() {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server: %v\n", err)
		}
	})

	log.Printf("Server listening on port %d\n", s.port)
	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		<-shutdownSignal
	} else {
		log.Printf("Error starting server: %v\n", err)
	}

	log.Println("Server has been shutdown")
}

func handleShutdown(onShutdownSignal func()) <-chan struct{} {
	shutdown := make(chan struct{})

	go func() {
		shutdownSignal := make(chan os.Signal, 1)
		signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
		<-shutdownSignal

		onShutdownSignal()

		close(shutdown)
	}()

	return shutdown
}
