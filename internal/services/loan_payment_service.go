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

func NewLoanPaymentService(
	db *gorm.DB,
	loanRepo repositories.LoanRepository,
	paymentRepo repositories.LoanPaymentRepository,
) LoanPaymentService {
	return &loanPaymentService{
		db:          db,
		loanRepo:    loanRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *loanPaymentService) MakePayment(paymentID int, loanID int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1) Get payment
		payment, err := s.paymentRepo.GetByID(paymentID)
		if err != nil {
			return err
		}
		if payment == nil {
			return errors.New("payment not found")
		}

		// 2) Validate payment belongs to loan
		if payment.LoanID != loanID {
			return errors.New("payment does not belong to the specified loan")
		}

		// 3) Prevent double pay
		if payment.Status == "paid" {
			return errors.New("payment already made")
		}

		// 4) Mark payment paid
		if err := s.paymentRepo.UpdateStatus(paymentID, "paid", time.Now()); err != nil {
			return err
		}

		// 5) Load loan (needed for TermsMonths)
		loan, err := s.loanRepo.GetByID(loanID)
		if err != nil {
			return err
		}
		if loan == nil {
			return errors.New("loan not found")
		}

		// 6) Check if all payments are paid
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

		// 7) If fully paid, close loan
		if paidCount == loan.TermsMonths {
			if err := s.loanRepo.UpdateStatus(loanID, "paid off"); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *loanPaymentService) ListPayments(loanID int) ([]models.LoanPayment, error) {
	return s.paymentRepo.ListByLoanID(loanID)
}
