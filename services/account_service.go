package services

import (
	"banking-api/internal/models"
	"banking-api/internal/repositories"

	"gorm.io/gorm"
)

type AccountService interface {
	CreateAccount(req *models.CreateAccountRequest, customerID, branchID int) (*models.Account, error)
	Transfer(fromAccountID, toAccountID int, amount float64) error
	Deposit(accountID int, amount float64) error
	GetStatement(accountID int) ([]models.Transaction, error)
}

type accountService struct {
	db *gorm.DB
	repo repositories.AccountRepository
	txRepo repositories.TransactionRepository
}

func NewAccountService(db *gorm.DB, repo repositories.AccountRepository, txRepo repositories.TransactionRepository) AccountService {
	return &accountService{db: db, repo: repo, txRepo: txRepo}
}

func (s *accountService) CreateAccount(req *models.CreateAccountRequest, customerID, branchID int) (*models.Account, error) {
	account := &models.Account{
		CustomerID: customerID,	
		BranchID:   branchID,
		Owner:	 req.Owner,
		Currency:   req.Currency,
		Balance:    0.0,
	}
	if err := s.repo.Create(account); err != nil {
		return nil, err
	}			
	return account, nil
}

func (s *accountService) Transfer(fromAccountID, toAccountID int, amount float64) error{
	return s.db.Transaction(func(tx *gorm.DB) error{
		fromAcc, err := s.repo.GetByID(fromAccountID)
		if err != nil {
			return err
		}

		if fromAcc.Balance < amount {
			return models.ErrInsufficientFunds
		}

		_,err=s.repo.GetByID(toAccountID)
		if err != nil {
			return err
		}

		if err:= s.repo.UpdateBalance(fromAccountID, fromAcc.Balance - amount); err != nil {
			return err
		}

		if err:= s.repo.UpdateBalance(toAccountID, fromAcc.Balance + amount); err != nil {
			return err
		}


		fromTx := &models.Transaction{
			fromAccountID: &fromAccountID,
			ToAccountID:   &toAccountID,
			Amount:        amount,
		}
		if err := s.txRepo.Create(fromTx); err != nil {
			return err
		}
	})
}