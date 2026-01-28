package handlers

import (
	"net/http"
	"strconv"

	"github.com/Mahesh252k/banking-api/internal/models"
	"github.com/Mahesh252k/banking-api/internal/repositories"
	"github.com/Mahesh252k/banking-api/internal/services"
	"github.com/Mahesh252k/banking-api/pkg/auth"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var dbConn *gorm.DB
var accountRepo repositories.AccountRepository
var txRepo repositories.TransactionRepository
var accountSvc services.AccountService
var loanRepo repositories.LoanRepository
var loanPaymentRepo repositories.LoanPaymentRepository
var loanSvc services.LoanService
var loanPaymentSvc services.LoanPaymentService

// InitHandlers initializes all handlers with database connection
func InitHandlers(db *gorm.DB) {
	dbConn = db

	accountRepo = repositories.NewAccountRepo(dbConn)
	txRepo = repositories.NewTransactionRepo(dbConn)
	accountSvc = services.NewAccountService(dbConn, accountRepo, txRepo)

	loanRepo = repositories.NewLoanRepo(dbConn)
	loanPaymentRepo = repositories.NewLoanPaymentRepo(dbConn)
	loanSvc = services.NewLoanService(dbConn, loanRepo, loanPaymentRepo)

	// correct order: (db, loanRepo, paymentRepo)
	loanPaymentSvc = services.NewLoanPaymentService(dbConn, loanRepo, loanPaymentRepo)
}

// -------------------- AUTH --------------------

func Register(c *gin.Context) {
	var req models.RegisterCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	customer := &models.Customer{
		Username:     req.Username,
		PasswordHash: string(hashed),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		Address:      req.Address,
	}

	if err := dbConn.Create(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateToken(customer.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

func Login(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var customer models.Customer
	if err := dbConn.Where("username = ?", loginReq.Username).First(&customer).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(customer.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// ACCOUNTS

func CreateAccount(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchID := 1
	account, err := accountSvc.CreateAccount(&req, userID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, account)
}

func ListAccounts(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// match your repo method name
	accounts, err := accountRepo.ListByCustomerID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func Transfer(c *gin.Context) {
	fromID, err := strconv.Atoi(c.Param("from_id"))
	if err != nil || fromID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_id"})
		return
	}

	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ToAccountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_account_id"})
		return
	}

	if err := accountSvc.Transfer(fromID, req.ToAccountID, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func Deposit(c *gin.Context) {
	accountID, err := strconv.Atoi(c.Param("account_id"))
	if err != nil || accountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}

	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := accountSvc.Deposit(accountID, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

// proper handler version (not service method)
func GetStatement(c *gin.Context) {
	accountID, err := strconv.Atoi(c.Param("id"))
	if err != nil || accountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}

	statement, err := accountSvc.GetStatement(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, statement)
}

// LOANS

func CreateLoan(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchID := 1
	loan, err := loanSvc.CreateLoan(&req, userID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loan)
}

func ListLoans(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	loans, err := loanSvc.ListLoans(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loans)
}

func MakePayment(c *gin.Context) {
	loanID, err := strconv.Atoi(c.Param("id"))
	if err != nil || loanID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	var req models.MakePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// service expects (paymentID, loanID)
	if err := loanPaymentSvc.MakePayment(req.PaymentID, loanID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment successful"})
}

func ListPayments(c *gin.Context) {
	loanID, err := strconv.Atoi(c.Param("id"))
	if err != nil || loanID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	payments, err := loanPaymentSvc.ListPayments(loanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// BENEFICIARIES

func AddBeneficiary(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req models.AddBeneficiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: save beneficiary to DB via service/repo
	c.JSON(http.StatusCreated, gin.H{"message": "beneficiary added successfully", "customer_id": userID})
}
