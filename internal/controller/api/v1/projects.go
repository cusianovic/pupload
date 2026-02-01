package v1

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/pupload/pupload/internal/controller/flows/service"
	"github.com/pupload/pupload/internal/controller/projects"
	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
)

func projectRoutes(f *service.FlowService, p *projects.ProjectService) http.Handler {
	r := chi.NewRouter()

	log := logging.ForService("project-api")
	r.Route("/{projectID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			projectID := chi.URLParam(r, "projectID")
			projectUUID, err := uuid.Parse(projectID)
			if err != nil {
				log.Error("invalid project id", "uuid", projectUUID.String())
				http.Error(w, "invalid project id", http.StatusBadRequest)
				return
			}

			proj, err := p.GetProject(r.Context(), projectUUID)
			if err != nil {
				log.Error("project id not found", "uuid", projectUUID.String())
				http.Error(w, "", http.StatusNotFound)
				return
			}

			render.JSON(w, r, proj)
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

			projectID := chi.URLParam(r, "projectID")
			var input models.Project

			if err := render.DecodeJSON(r.Body, &input); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}

			if input.ID.String() != projectID {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			if err := p.SaveProject(r.Context(), input); err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			render.Status(r, 200)
		})

		r.Route("/flows", func(r chi.Router) {
			r.Route("/{flowName}", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					projectID := chi.URLParam(r, "projectID")
					projectUUID, err := uuid.Parse(projectID)
					if err != nil {
						http.Error(w, "invalid project id", http.StatusBadRequest)
						return
					}

					flowName := chi.URLParam(r, "flowName")
					if flowName == "" {
						http.Error(w, "no flow name specified", http.StatusBadRequest)
						return
					}

					bundle, err := p.GetFlowFromProject(r.Context(), projectUUID, flowName)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}

					run, err := f.RunFlow(bundle.Flow, bundle.Defs)
					if err != nil {
						http.Error(w, fmt.Sprintf("unable to run flow: %s", err), http.StatusInternalServerError)
						return
					}

					render.JSON(w, r, run)
				})
			})

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {})
			r.Patch("/", func(w http.ResponseWriter, r *http.Request) {})
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {})

			r.Post("/validate", func(w http.ResponseWriter, r *http.Request) {})
			r.Post("/run", func(w http.ResponseWriter, r *http.Request) {})
		})

	})

	return r
}
