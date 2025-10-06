package db

import (
	models "emplacc-api/internal/domain"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB_conn *gorm.DB = GetDBConnection()

func GetDBConnection() *gorm.DB {
	time.Sleep(5 * time.Second)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")
	timezone := os.Getenv("DB_TIMEZONE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, pass, name, port, sslMode, timezone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	db.InstanceSet("gorm:cache:prepared_statement", nil)
	db.InstanceSet("gorm:cache:schema", nil)

	// включаем каскадные FK
	db = db.Set("gorm:foreignKeyConstraints", true)

	// берем *sql.DB для raw запросов
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get *sql.DB from GORM:", err)
	}

	// пул соединений
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
	// 1. Базовые сущности (без внешних ключов или с необязательными)
		&models.User{},
		&models.Role{},
		&models.Team{},
		&models.Problem{},

		// 2. Джойн-таблицы и сущности, ссылающиеся на базовые
		&models.UserRole{},
		&models.TeamMember{},
		&models.ForumMessage{}, // ссылается на Problem и User
		&models.Project{},      // ссылается на User (CreatedBy)

		// 3. Сущности второго уровня
		&models.Board{},        // ссылается на Project

		// 4. Статусы — ссылаются на Board
		&models.Status{},

		// 5. Задачи — ссылаются на Status, User (AssignedTo, CreatedBy)
		&models.Task{},

		// 6. Отчёты и связанные с задачами/пользователями
		&models.Attendance{},
		&models.DailyReport{},
		&models.ReportProblem{},
		&models.HelpRequest{},
		&models.CompletedWork{},
		&models.TomorrowPlans{},
		&models.ProjectTeam{},  // ссылается на Project и Team
		&models.Subscription{},
	); err != nil {
		log.Fatal("AutoMigrate failed:", err)
	}

	return db
}
