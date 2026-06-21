package ports

import (
	"time"

	models "emplacc-api/internal/domain"

	"github.com/google/uuid"
)

type GitRepository interface {
	CreateRepository(repo *models.CodeRepository) error
	GetRepository(id uuid.UUID) (*models.CodeRepository, error)
	ListRepositoriesByProject(projectID uuid.UUID) ([]models.CodeRepository, error)
	SoftDeleteRepository(id uuid.UUID) (bool, error)
	TouchRepositorySync(id uuid.UUID, at time.Time) error
	UpsertCommits(commits []models.Commit) (created int, err error)
	LinkCommitToTask(commitID uuid.UUID, taskID *uuid.UUID) (bool, error)
	ListCommitsByTask(taskID uuid.UUID, limit, offset int) ([]models.Commit, int64, error)
	ListCommitsByProject(projectID uuid.UUID, limit, offset int) ([]models.Commit, int64, error)
	ListCommitsByRepository(repoID uuid.UUID, limit, offset int) ([]models.Commit, int64, error)
}
