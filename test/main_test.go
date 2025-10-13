package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"emplacc-api/internal/controller"
	models "emplacc-api/internal/domain"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	time.Sleep(5 * time.Second)

	// Создаём тестовую БД один раз перед всеми тестами
	createTestDatabase()

	// Запускаем все тесты
	code := m.Run()

	// (опционально) можно удалить БД после — но обычно не нужно
	// dropTestDatabase()

	os.Exit(code)
}

var (
	origDBConn    = controller.DBConn
	origAuthorize = controller.Authorize
)

func createTestDatabase() {
    // Используем единый подход для Docker и локальной среды
    port := getEnv("TEST_DB_PORT", "5432")
    user := getEnv("TEST_DB_USER", "postgres")
    password := getEnv("TEST_DB_PASSWORD", "admin")
    dbname := getEnv("TEST_DB_NAME", "postgres")

    // Подключаемся к служебной БД
    adminDSN := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable", 
         user, password, dbname, port)
    adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
    if err != nil {
        panic("Failed to connect to admin DB: " + err.Error())
    }

    sqlDB, err := adminDB.DB()
    if err != nil {
        panic("Failed to get *sql.DB: " + err.Error())
    }
    defer sqlDB.Close()

    // Создаем тестовую БД если не существует
    _, err = sqlDB.Exec("CREATE DATABASE test")
    if err != nil && !strings.Contains(err.Error(), "already exists") {
        panic("Failed to create test DB: " + err.Error())
    }
}

func setupTestDB(t *testing.T) *gorm.DB {
    host := getEnv("TEST_DB_HOST", "localhost")
    port := getEnv("TEST_DB_PORT", "5432")
    user := getEnv("TEST_DB_USER", "postgres")
    password := getEnv("TEST_DB_PASSWORD", "admin")

    testDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=test port=%s sslmode=disable TimeZone=UTC", 
        host, user, password, port)
    
    db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
    db = db.Session(&gorm.Session{Logger: logger.Discard})
    require.NoError(t, err)

    // Очищаем схему
    sqlDB, err := db.DB()
    require.NoError(t, err)
    _, err = sqlDB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
    require.NoError(t, err)

    // Миграции
    err = db.AutoMigrate(
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
	)
    require.NoError(t, err)

    return db
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}