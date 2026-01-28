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

func calculateEMI(prinicipal, monthlyRate float64, months int) float64 {
	if months == 0 {
		return 0
	}
	r := monthlyRate / 12

	if r == 0 {
		return prinicipal / float64(months)
	}

	power := math.Pow(1+r, float64(months))
	emi := prinicipal * r * power / (power - 1)
	return emi
}

func (s *loanService) CreateLoan(req *models.CreateLoanRequest, customerID, branchID int) (*models.Loan, error) {
	var loan *models.Loan
	err := s.db.Transaction(func(tx *gorm.DB) error {
		//calculate total payable
		monthlyRate := req.InterestRate
		emi := calculateEMI(req.Amount, monthlyRate, req.TermsMonths)
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
		//generate payment (equal EMI)
		for i := 0; i < req.TermsMonths; i++ {
			dueDate := time.Now().AddDate(0, i, 0) //monthly
			payment := &models.LoanPayment{
				LoanID:  loan.ID,
				Amount:  emi,
				DueDate: dueDate,
				Status:  "pending",
			}
			if err := s.paymentRepo.Create(payment); err != nil {
				return err // rollback on failure
			}
		}

		//Reload with payments
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
