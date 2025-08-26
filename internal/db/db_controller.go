package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB_conn *gorm.DB = getDBConnection()

func getDBConnection() *gorm.DB {
	time.Sleep(5 * time.Second)
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")
	timezone := os.Getenv("DB_TIMEZONE")

	connectData := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, pass, name, port, sslMode, timezone,
	)

	db, err := gorm.Open(postgres.Open(connectData), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	// Включаем поддержку внешних ключей с каскадом
	db = db.Set("gorm:foreignKeyConstraints", true)

	return db
}
