package v1

import (
	"net/http"

	flows "github.com/pupload/pupload/internal/controller/flows/service"
	"github.com/pupload/pupload/internal/controller/projects"

	"github.com/go-chi/chi/v5"
)

func HandleAPIRoutes(f *flows.FlowService, p *projects.ProjectService) http.Handler {

	r := chi.NewRouter()

	r.Mount("/flow", handleFlowRoutes(f))
	r.Mount("/upload", handleUploadRoutes())
	r.Mount("/project", projectRoutes(f, p))

	return r

}
