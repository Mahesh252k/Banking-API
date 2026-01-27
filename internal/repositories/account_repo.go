package repositories

import (
	"banking-api/internal/models"

	"gorm.io/gorm"
)

type AccountRepository interface {
	Create(account *models.Account) error
	GetByID(id int) (*models.Account, error)
	UpdateBalance(account *models.Account) error
	ListByCustomerID(customerID int) ([]models.Account, error)
}

func NewAccountRepo(db *gorm.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *accountRepo) GetByID(id int) (*models.Account, error) {
	var account models.Account
	if err := r.db.Preload("Customer").Preload("Branch").First(&account, id).Error; err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *accountRepo) UpdateBalance(account *models.Account) error {
	return r.db.Model(account).Update("balance", account.Balance).Error
}

func (r *accountRepo) ListByCustomerID(customerID int) ([]models.Account, error) {
	var accounts []models.Account
	if err := r.db.Preload("Customer").Preload("Branch").Where("customer_id = ?", customerID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
