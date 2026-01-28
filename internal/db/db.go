package db

import (
	"log"
	"os"

	"github.com/Mahesh252k/banking-api/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not loaded. Ensure .env is loaded")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// ✅ Auto-create/update tables based on your models (dev-friendly)
	if err := db.AutoMigrate(
		&models.Customer{},
		&models.Branch{},
		&models.Account{},
		&models.Transaction{},
		&models.Loan{},
		&models.LoanPayment{},
		&models.Beneficiary{},
	); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	log.Println("Database connected successfully")
	return db
}
