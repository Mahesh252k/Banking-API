package handlers

import (
	"banking-api/internal/models"
	"net/http"
	"strconv"

	"github.com/Mahesh252k/banking-api/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"repositories"
	"services"
)

var dbConn *gorm.DB
var accountRepo repositories.AccountRepository
var txRepo repositories.TransactionRepository
var accountSvc services.AccountService
var loanRepo repositories.LoanRepository
var loanPaymentRepo repositories.LoanPaymentRepository
var loanSvc services.LoanService
var loanPaymentSvc services.LoanPaymentService

// InitHandlers initialize all handlers with database connection
func InitHandlers(db *gorm.DB) {
	dbConn = db

	//Init repo and services
	accountRepo = repositories.NewAccountRepo(dbConn)
	txRepo = repositories.NewTransactionRepo(dbConn)
	accountSvc = services.NewAccountService(dbConn, accountRepo, txRepo)

	loanRepo = repositories.NewLoanRepo(dbConn)
	loanPaymentRepo = repositories.NewLoanPaymentRepo(dbConn)
	loanSvc = services.NewLoanService(dbConn, loanRepo, loanPaymentRepo)
	loanPaymentSvc = services.NewLoanPaymentService(dbConn, loanPaymentRepo, loanRepo)
}

// Auth
func Register(c *gin.Context) {
	var req models.RegisterCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
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

	token, _ := auth.GenerateToken(customer.ID)
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(loginReq.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := auth.GenerateToken(customer.ID)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Accounts
func CreateAccount(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	accounts, err := accountRepo.ListByCustomer(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

//Transfer

func Transfer(c *gin.Context) {
	fromID, err := strconv.Atoi(c.Param("from_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_id"})
		return
	}

	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ToAccountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to account_id"})
		return
	}

	if err := accountSvc.Transfer(fromID, req.ToAccountID, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

// Deposits
func Deposit(c *gin.Context) {
	accountID, err := strconv.Atoi(c.Param("account_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id"})
		return
	}

	if accountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id"})
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
	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
}

func (s *AccountService) GetStatement(accountID int) ([]models.Transaction, error) {
	accountID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id"})
		return nil, err
	}
	statement, err := accountSvc.GetStatement(accountID)
	c.JSON(http.StatusOK, statement)
}

// Loans
func CreateLoan(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
