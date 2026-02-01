package projects

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
)

type ProjectService struct {
	repo ProjectRepo
	log  *slog.Logger
}

type ProjectBundle struct {
	Flow models.Flow
	Defs []models.NodeDef
}

func CreateProjectService(cfg RedisProjectRepoConfig) (*ProjectService, error) {
	repo, err := CreateRedisProjectRepo(cfg)
	return &ProjectService{
		repo: repo,
		log:  logging.ForService("project"),
	}, err
}

func (p *ProjectService) Close() {
	p.repo.Close()
}

func (p *ProjectService) SaveProject(ctx context.Context, project models.Project) error {
	return p.repo.SaveProject(ctx, project)
}

func (p *ProjectService) GetProject(ctx context.Context, projectID uuid.UUID) (models.Project, error) {
	proj, err := p.repo.LoadProject(ctx, projectID)
	if err != nil {
		if err != ErrProjectDoesNotExist {
			p.log.Error("error loading project", "err", err)
		}

		return proj, err
	}

	return proj, nil
}

func (p *ProjectService) GetFlowFromProject(ctx context.Context, projectID uuid.UUID, flowName string) (*ProjectBundle, error) {
	proj, err := p.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var b *ProjectBundle
	for _, flow := range proj.Flows {
		if flowName == flow.Name {
			b = &ProjectBundle{
				Flow: flow,
			}
		}
	}

	if b == nil {
		return nil, fmt.Errorf("Flow %s does not exist in project %s", flowName, projectID.String())
	}

	b.Defs = proj.NodeDefs
	return b, nil

}
