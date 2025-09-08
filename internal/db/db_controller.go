package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB_conn *gorm.DB = getDBConnection()

func getDBConnection() *gorm.DB {
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

	// Включаем поддержку внешних ключей с каскадом
	db = db.Set("gorm:foreignKeyConstraints", true)

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get *sql.DB from GORM:", err)
	}

	// Настройки пула
	sqlDB.SetMaxOpenConns(50)           // максимум открытых соединений
	sqlDB.SetMaxIdleConns(10)           // максимум неиспользуемых соединений
	sqlDB.SetConnMaxLifetime(time.Hour) // максимальное время жизни соединения

	return db
}
