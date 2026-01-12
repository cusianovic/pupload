package controllerserver

import (
	"net/http"
	v1 "pupload/internal/controller/api/v1"
	config "pupload/internal/controller/config"
	flows "pupload/internal/controller/flows/service"
	"pupload/internal/logging"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(config config.ControllerSettings, f *flows.FlowService) http.Handler {

	log := logging.ForService("server")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/api/v1", v1.HandleAPIRoutes(f))

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Info("Route", "method", method, "route", route)
		return nil
	}

	chi.Walk(r, walkFunc)

	return r
}
