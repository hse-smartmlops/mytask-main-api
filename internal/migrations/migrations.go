package migrations

import (
	"log/slog"

	"github.com/go-gormigrate/gormigrate/v2"

	models "emplacc-api/internal/domain/models"

	"gorm.io/gorm"
)

var baseEntities = []interface{}{
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

var schemaMigrations = []*gormigrate.Migration{
	{
		ID: "202409150001_initial_schema",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(baseEntities...); err != nil {
				return err
			}

			for _, idx := range []struct {
				model interface{}
				name  string
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
			} {
				if !tx.Migrator().HasIndex(idx.model, idx.name) {
					if err := tx.Migrator().CreateIndex(idx.model, idx.name); err != nil {
						return err
					}
				}
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(
				&models.ReportProblem{},
				&models.TomorrowPlans{},
				&models.CompletedWork{},
				&models.HelpRequest{},
				&models.DailyReport{},
				&models.Attendance{},
				&models.Task{},
				&models.Status{},
				&models.Board{},
				&models.ProjectTeam{},
				&models.TeamMember{},
				&models.Team{},
				&models.Subscription{},
				&models.ForumMessage{},
				&models.Problem{},
				&models.Project{},
				&models.UserRole{},
				&models.Role{},
				&models.User{},
			)
		},
	},
	{
		ID: "202409150002_storage_columns",
		Migrate: func(tx *gorm.DB) error {
			if !tx.Migrator().HasColumn(&models.User{}, "avatar_path") {
				if err := tx.Migrator().AddColumn(&models.User{}, "AvatarPath"); err != nil {
					return err
				}
			}
			if !tx.Migrator().HasColumn(&models.DailyReport{}, "storage_object") {
				if err := tx.Migrator().AddColumn(&models.DailyReport{}, "StorageObject"); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if tx.Migrator().HasColumn(&models.DailyReport{}, "storage_object") {
				if err := tx.Migrator().DropColumn(&models.DailyReport{}, "StorageObject"); err != nil {
					return err
				}
			}
			if tx.Migrator().HasColumn(&models.User{}, "avatar_path") {
				if err := tx.Migrator().DropColumn(&models.User{}, "AvatarPath"); err != nil {
					return err
				}
			}
			return nil
		},
	},
}

func Migrate(db *gorm.DB, logger *slog.Logger) error {
	mg := gormigrate.New(db, gormigrate.DefaultOptions, schemaMigrations)
	if err := mg.Migrate(); err != nil {
		logger.Error("database migrate failed", slog.String("error", err.Error()))
		return err
	}
	logger.Info("database migrations applied")
	return nil
}

func Rollback(db *gorm.DB, logger *slog.Logger) error {
	mg := gormigrate.New(db, gormigrate.DefaultOptions, schemaMigrations)
	if err := mg.RollbackLast(); err != nil {
		logger.Error("database rollback failed", slog.String("error", err.Error()))
		return err
	}
	logger.Info("database migration rolled back")
	return nil
}
