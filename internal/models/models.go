package models

import (
	"time"
)

type Customer struct {
	ID           int       `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	Username     string    `gorm:"unique;size:50" json:"username"`
	PasswordHash string    `json:"-" gorm:"size:255"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `gorm:"unique;size:100" json:"email"`
	Phone        string    `gorm:"size:20" json:"phone"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`

	Accounts      []Account     `gorm:"foreignKey:CustomerID" json:"-"`
	Loans         []Loan        `gorm:"foreignKey:CustomerID" json:"-"`
	Beneficiaries []Beneficiary `gorm:"foreignKey:CustomerID" json:"-"`
}

type Branch struct {
	ID      int    `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	Name    string `json:"name"`
	Code    string `gorm:"unique;size:10" json:"code"`
	City    string `gorm:"size:50" json:"city"`
	Address string `json:"address"`
	Phone   string `json:"phone"`

	Accounts []Account `gorm:"foreignKey:BranchID" json:"-"`
	Loans    []Loan    `gorm:"foreignKey:BranchID" json:"-"`
}

type Account struct {
	ID           int           `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	CustomerID   int           `json:"customer_id" gorm:"type:int;index"`
	Customer     *Customer     `gorm:"foreignKey:CustomerID" json:"customer"`
	BranchID     int           `json:"branch_id" gorm:"type:int;index"`
	Branch       *Branch       `gorm:"foreignKey:BranchID" json:"branch"`
	Owner        string        `json:"owner"`
	Balance      float64       `gorm:"type:decimal(15,2)" json:"balance"`
	Currency     string        `json:"currency"`
	CreatedAt    time.Time     `json:"created_at"`
	Transactions []Transaction `gorm:"foreignKey:FromAccountID;references:ID" json:"-"`
}

type Transaction struct {
	ID            int       `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	FromAccountID *int      `json:"from_account_id" gorm:"type:int;index"`
	ToAccountID   *int      `json:"to_account_id" gorm:"type:int;index"`
	LoanPaymentID *int      `json:"loan_payment_id" gorm:"type:int;index"`
	BeneficiaryID *int      `json:"beneficiary_id" gorm:"type:int;index"`
	Amount        float64   `gorm:"type:decimal(15,2)" json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}

type Loan struct {
	ID           int           `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	CustomerID   int           `json:"customer_id" gorm:"type:int;index"`
	Customer     *Customer     `gorm:"foreignKey:CustomerID" json:"customer"`
	BranchID     int           `json:"branch_id" gorm:"type:int;index"`
	Branch       *Branch       `gorm:"foreignKey:BranchID" json:"branch"`
	Amount       float64       `gorm:"type:decimal(15,2)" json:"amount"`
	InterestRate float64       `gorm:"type:decimal(5,2)" json:"interest_rate"`
	TermsMonths  int           `json:"terms_months"`
	TotalPayable float64       `gorm:"type:decimal(15,2)" json:"total_payable"`
	Status       string        `json:"status"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      time.Time     `json:"end_date"`
	CreatedAt    time.Time     `json:"created_at"`
	Payments     []LoanPayment `gorm:"foreignKey:LoanID" json:"payments"`
}

type LoanPayment struct {
	ID        int       `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	LoanID    int       `json:"loan_id" gorm:"type:int;index"`
	Loan      *Loan     `gorm:"foreignKey:LoanID" json:"loan"`
	Amount    float64   `gorm:"type:decimal(15,2)" json:"amount"`
	DueDate   time.Time `json:"due_date"`
	PaidDate  time.Time `json:"paid_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateLoanRequest struct {
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	InterestRate float64 `json:"interest_rate" binding:"required,gt=0"`
	TermsMonths  int     `json:"terms_months" binding:"required,gt=0"`
}

type Beneficiary struct {
	ID            int    `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	CustomerID    int    `json:"customer_id" gorm:"type:int;index"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
}

type CreateAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

type TransferRequest struct {
	ToAccountID int     `json:"to_account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
}

type RegisterCustomerRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}

type AddBeneficiaryRequest struct {
	Name          string `json:"name" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	BankName      string `json:"bank_name"`
}

type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type MakePaymentRequest struct {
	PaymentID int `json:"payment_id" binding:"required"`
}
