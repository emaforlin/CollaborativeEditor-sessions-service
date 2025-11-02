package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emaforlin/ce-sessions-service/config"
)

const gracefulShutdownTimeout = 30 * time.Second

type httpServer struct {
	httpServer *http.Server
	mux        *http.ServeMux
}

type APIServer struct {
	config     config.Config
	httpServer *httpServer
}

func (s *APIServer) GetHTTPServer() *http.Server {
	return s.httpServer.httpServer
}

func (s *APIServer) RegisterHandler(pattern string, handler http.HandlerFunc) {
	log.Printf("Register handler route %s", pattern)
	s.httpServer.mux.HandleFunc(pattern, handler)
}

func (s *APIServer) RegisterHandlerWithMiddlewares(pattern string, handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	s.RegisterHandler(pattern, finalHandler)
}

func (s *APIServer) Start() error {
	go func() {
		log.Printf("Starting server on %s", s.config.GetServerAddress())
		if err := s.GetHTTPServer().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests a deadline to complete
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.GetHTTPServer().Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited")
	return nil
}

func NewAPIServer(cfg config.Config) *APIServer {
	mux := http.NewServeMux()
	return &APIServer{
		config: cfg,
		httpServer: &httpServer{
			mux: mux,
			httpServer: &http.Server{
				Addr:    cfg.GetServerAddress(),
				Handler: mux,
			},
		},
	}
}
