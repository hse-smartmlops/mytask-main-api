package migrations

import (
	"log/slog"

	models "emplacc-api/internal/domain/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, logger *slog.Logger) error {
	entities := []interface{}{
		&models.User{},
		&models.Role{},
		&models.Team{},
		&models.Project{},
		&models.Board{},
		&models.Status{},
		&models.Task{},
		&models.Attendance{},
		&models.DailyReport{},
		&models.HelpRequest{},
		&models.CompletedWork{},
		&models.TomorrowPlans{},
		&models.Problem{},
		&models.ForumMessage{},
		&models.Subscription{},
		&models.ProjectTeam{},
		&models.TeamMember{},
		&models.UserRole{},
		&models.ReportProblem{},
	}

	if err := db.AutoMigrate(entities...); err != nil {
		logger.Error("auto migrate failed", slog.String("error", err.Error()))
		return err
	}

	indexCreations := []struct {
		value interface{}
		field string
	}{
		{&models.Board{}, "ProjectID"},
		{&models.Status{}, "BoardID"},
		{&models.Task{}, "StatusID"},
		{&models.Task{}, "AssignedTo"},
		{&models.Attendance{}, "UserID"},
		{&models.ReportProblem{}, "ReportID"},
		{&models.ReportProblem{}, "ProblemID"},
		{&models.TeamMember{}, "TeamID"},
		{&models.TeamMember{}, "UserID"},
		{&models.ProjectTeam{}, "ProjectID"},
		{&models.ProjectTeam{}, "TeamID"},
	}

	for _, idx := range indexCreations {
		if db.Migrator().HasIndex(idx.value, idx.field) {
			continue
		}
		if err := db.Migrator().CreateIndex(idx.value, idx.field); err != nil {
			logger.Error("create index failed", slog.String("index", idx.field), slog.String("error", err.Error()))
			return err
		}
	}

	logger.Info("database migrations applied")
	return nil
}
