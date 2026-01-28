package services

import (
	"math"
	"time"

	"github.com/Mahesh252k/banking-api/internal/models"
	"github.com/Mahesh252k/banking-api/internal/repositories"
	"gorm.io/gorm"
)

type LoanService interface {
	CreateLoan(req *models.CreateLoanRequest, customerID, branchID int) (*models.Loan, error)
	ListLoans(customerID int) ([]models.Loan, error)
	UpdateStatus(loanID int, status string) error
}

type loanService struct {
	db          *gorm.DB
	loanRepo    repositories.LoanRepository
	paymentRepo repositories.LoanPaymentRepository
}

func NewLoanService(db *gorm.DB, loanRepo repositories.LoanRepository, paymentRepo repositories.LoanPaymentRepository) LoanService {
	return &loanService{db: db, loanRepo: loanRepo, paymentRepo: paymentRepo}
}

// annualRate is like 10 for 10%
func calculateEMI(principal, annualRate float64, months int) float64 {
	if months <= 0 {
		return 0
	}
	monthlyRate := (annualRate / 100) / 12
	if monthlyRate == 0 {
		return principal / float64(months)
	}
	power := math.Pow(1+monthlyRate, float64(months))
	return principal * monthlyRate * power / (power - 1)
}

func (s *loanService) CreateLoan(req *models.CreateLoanRequest, customerID, branchID int) (*models.Loan, error) {
	var loan *models.Loan

	err := s.db.Transaction(func(tx *gorm.DB) error {
		emi := calculateEMI(req.Amount, req.InterestRate, req.TermsMonths)
		totalPayable := emi * float64(req.TermsMonths)

		loan = &models.Loan{
			CustomerID:   customerID,
			BranchID:     branchID,
			Amount:       req.Amount,
			InterestRate: req.InterestRate,
			TermsMonths:  req.TermsMonths,
			TotalPayable: totalPayable,
			Status:       "approved",
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, req.TermsMonths, 0),
		}

		if err := s.loanRepo.Create(loan); err != nil {
			return err
		}

		for i := 1; i <= req.TermsMonths; i++ {
			dueDate := time.Now().AddDate(0, i, 0)
			payment := &models.LoanPayment{
				LoanID:  loan.ID,
				Amount:  emi,
				DueDate: dueDate,
				Status:  "pending",
			}
			if err := s.paymentRepo.Create(payment); err != nil {
				return err
			}
		}

		fullLoan, err := s.loanRepo.GetByID(loan.ID)
		if err != nil {
			return err
		}
		loan = fullLoan
		return nil
	})

	if err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *loanService) ListLoans(customerID int) ([]models.Loan, error) {
	return s.loanRepo.ListByCustomerID(customerID)
}

func (s *loanService) UpdateStatus(loanID int, status string) error {
	return s.loanRepo.UpdateStatus(loanID, status)
}
