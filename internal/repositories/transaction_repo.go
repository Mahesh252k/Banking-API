package repositories

import (
	"banking-api/internal/models"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(transaction *models.Transaction) error
}
type transactionRepo struct {
	db *gorm.DB
}

func NewTransactionRepo(db *gorm.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}
