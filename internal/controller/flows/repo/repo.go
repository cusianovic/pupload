package repo

import (
	"context"
	"pupload/internal/controller/flows/runtime"
	"pupload/internal/models"
)

type ProjectRepo interface {
	SaveProject(ctx context.Context, project models.Project) error
	LoadProject(ctx context.Context, tenantID, projectName string) (models.Project, error)
	DeleteProject(ctx context.Context, tenantID, projectName string) error
	Close(ctx context.Context) error
}

type RuntimeRepo interface {
	SaveRuntime(rt runtime.RuntimeFlow) error
	LoadRuntime(runID string) (runtime.RuntimeFlow, error)
	DeleteRuntime(runID string) error
	ListRuntimeIDs() ([]string, error)
	Close(ctx context.Context) error
}
