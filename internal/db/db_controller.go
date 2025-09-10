package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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

	// инициализация через init.sql
	initSQL(sqlDB)

	return db
}

func initSQL(db *sql.DB) {
	script, err := ioutil.ReadFile("init.sql") // файл лежит рядом с бинарником
	if err != nil {
		log.Fatalf("Ошибка чтения init.sql: %v", err)
	}

	// разбиваем по `;`, чтобы выполнить по отдельности
	queries := strings.Split(string(script), ";")
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		_, err := db.Exec(q)
		if err != nil {
			log.Fatalf("Ошибка выполнения запроса [%s]: %v", q, err)
		}
	}

	log.Println("init.sql успешно выполнен")
}
