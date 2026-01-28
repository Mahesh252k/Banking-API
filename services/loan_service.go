package services

import (
	"banking-api/internal/models"
	"banking-api/internal/repositories"
	"gorm.io/gorm"
	"time"
	"math"
)

type LoanService interface {
	CreateLoan(req *models.createLoanRequest, customerID, branchID int)(*models.Loan, error)
	ListLoans(customerID int)([]models.Loan, error)