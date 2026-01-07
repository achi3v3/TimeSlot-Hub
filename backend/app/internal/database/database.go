package database

import (
	"app/pkg/models"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *Database
	once sync.Once
)

// Database represents database connection structure
type Database struct {
	DB   *gorm.DB
	name string
}

// Init initializes database connection and runs migrations
func Init() *Database {
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	dbname := GetEnv("DB_NAME", "postgres")
	user := GetEnv("DB_USER", "postgres")
	password := GetEnv("DB_PASSWORD", "password")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		log.Fatal("Not all required database environment variables are set")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)

	log.Printf("Connecting to database: host=%s, port=%s, dbname=%s, user=%s", host, port, dbname, user)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Successfully connected to database")
	db.AutoMigrate(&models.User{}, &models.UserRole{}, &models.Slot{}, &models.Record{}, &models.Notification{}, &models.AdClickStats{})
	return &Database{
		name: "database",
		DB:   db,
	}
}

// GetDB returns initialized database connection
func GetDB() *Database {
	if db == nil {
		sleep := 1 * time.Second
		once.Do(func() {
			for db == nil {
				sleep = sleep * 2
				fmt.Printf("Database is unavailable. Wait for %d sec.\n", sleep)
				time.Sleep(sleep)
				db = Init()
			}
		})
	}
	return db
}

func (d *Database) Shutdown(ctx context.Context) error {
	if d.DB == nil {
		return nil
	}
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return sqlDB.Close()
}
func (d *Database) Name() string {
	return d.name
}
func GetEnv(key, defaultValue string) string {
	if err := godotenv.Load("/app/.env"); err != nil {
		log.Printf("Notice: Could not load .env file from /app/.env: %v", err)
	}
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Notice: Could not load .env file from /.env: %v", err)
	}
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
