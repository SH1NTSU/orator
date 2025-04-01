package application

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/SH1NTSU/orator/handlers"
)

type App struct {
	router http.Handler
}

func New() *App {
	handlers.InitWorkerPool(5)

	return &App{
		router: loadRoutes(),
	}
}

func (a *App) Start(ctx context.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: a.router,
	}

	fmt.Println("Server Listening on http://localhost:" + port)
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
