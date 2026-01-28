package db

import (
	"log"

	"github.com/Mahesh252k/banking-api/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	dsn := ""

	if dsn == "" {
		log.Fatal("DB_DSN not loaded. Ensure .env is loaded")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("[Error] failed to intialize database, got error %v", err)
	}

	//Auto-Migrate models
	err = db.AutoMigrate(&models.Customer{}, &models.Branch{}, &models.Account{}, &models.Transaction{}, &models.LoanPayment{}, &models.AddBeneficiaryRequest{}, &models.Loan{})
	if err != nil {
		log.Fatalf("Automigrated failed: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return db
}
