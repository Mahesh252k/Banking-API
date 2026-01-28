package services

import (
	"errors"
	"time"

	"github.com/Mahesh252k/banking-api/internal/models"
	"github.com/Mahesh252k/banking-api/internal/repositories"

	"gorm.io/gorm"
)

type LoanPaymentService interface {
	MakePayment(paymentID int, loanID int) error
	ListPayments(loanID int) ([]models.LoanPayment, error)
}

type loanPaymentService struct {
	db          *gorm.DB
	loanRepo    repositories.LoanRepository
	paymentRepo repositories.LoanPaymentRepository
}

func NewLoanPaymentService(db *gorm.DB, loanRepo repositories.LoanRepository, paymentRepo repositories.LoanPaymentRepository) LoanPaymentService {
	return &loanPaymentService{db: db, loanRepo: loanRepo, paymentRepo: paymentRepo}
}

func (s *loanPaymentService) MakePayment(paymentID int, loanID int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		payment, err := s.paymentRepo.GetByID(paymentID)
		if err != nil {
			return errors.New("payment not found")
		}

		if payment.LoanID != loanID {
			return errors.New("payment does not belong to the specified loan")
		}

		if payment.Status == "paid" {
			return errors.New("payment already made")
		}

		//Update Payment
		if err := s.paymentRepo.UpdateStatus(paymentID, "paid", time.Now()); err != nil {
			return err
		}

		//Update LOAN IF All Paid
		payments, err := s.paymentRepo.ListByLoanID(loanID)
		if err != nil {
			return err
		}

		paidCount := 0
		for _, p := range payments {
			if p.Status == "paid" {
				paidCount++
			}
		}

		if paidCount == loan.TermsMonths {
			s.loanRepo.UpdateStatus(loanID, "paid off")
		}

		return nil
	})
}

func (s *loanPaymentService) ListPayments(loanID int) ([]models.LoanPayment, error) {
	return s.paymentRepo.ListByLoanID(loanID)
}
