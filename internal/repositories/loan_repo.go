package repositories

import (
	"github.com/Mahesh252k/banking-api/internal/models"
	"gorm.io/gorm"
)

type LoanRepository interface {
	Create(loan *models.Loan) error
	GetByID(id int) (*models.Loan, error)
	ListByCustomerID(customerID int) ([]models.Loan, error)
	UpdateStatus(id int, status string) error
}

type loanRepo struct {
	db *gorm.DB
}

func NewLoanRepo(db *gorm.DB) LoanRepository {
	return &loanRepo{db: db}
}

func (r *loanRepo) Create(loan *models.Loan) error {
	return r.db.Create(loan).Error
}

func (r *loanRepo) GetByID(id int) (*models.Loan, error) {
	var loan models.Loan
	if err := r.db.Preload("Customer").Preload("Payments").First(&loan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &loan, nil
}

func (r *loanRepo) ListByCustomerID(customerID int) ([]models.Loan, error) {
	var loans []models.Loan
	if err := r.db.Preload("Customer").Where("customer_id = ?", customerID).Find(&loans).Error; err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *loanRepo) UpdateStatus(id int, status string) error {
	return r.db.Model(&models.Loan{}).Where("id = ?", id).Update("status", status).Error
}
