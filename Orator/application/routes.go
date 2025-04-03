package application

import (
	"net/http"

	"github.com/SH1NTSU/orator/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func loadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to MajsterApp API"))
	})

	router.Route("/api/v1", loadHandlerRoutes)

	return router

}

func loadHandlerRoutes(router chi.Router) {
	Handler := &handlers.Order{}
	router.Post("/to_speech", Handler.ToSpeach)
	router.Patch("/set_parameter", Handler.SetParameter)
}
