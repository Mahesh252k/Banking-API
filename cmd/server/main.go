package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Mahesh252k/banking-api/internal/db"
	"github.com/Mahesh252k/banking-api/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("cannot get working directory:", err)
	}

	// Always load .env from project root (2 levels up from cmd/server)
	projectRoot := filepath.Clean(filepath.Join(dir, "..", ".."))
	envPath := filepath.Join(projectRoot, ".env")

	log.Printf("Loading environment variables from %s", envPath)

	if err := godotenv.Overload(envPath); err != nil {
		log.Printf("warning: could not load .env from %s: %v", envPath, err)
	} else {
		log.Println("successfully loaded .env file")
	}

	dsn := os.Getenv("DB_DSN")
	log.Printf("DB user check: %s", dsn) // TEMP debug (remove later)

	jwtSecret := os.Getenv("JWT_SECRET")
	port := os.Getenv("PORT")

	if dsn == "" {
		log.Fatal("DB_DSN is not loaded")
	}

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not loaded")
	}

	if port == "" {
		port = "8080"
	}

	dbConn := db.Connect()
	handlers.InitHandlers(dbConn)
	r := gin.Default()
	r.Use(gin.Logger()) //logging middleware

	//CORS Middleware - applied globally for all routes
	r.Use(corsMiddleware())

	//public routes
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)

	//protected routes
	protected := r.Group("")
	protected.Use(authMiddleware())

	//accounts
	protected.POST("/accounts", handlers.CreateAccount)
	protected.GET("/accounts", handlers.ListAccounts)
	protected.POST("/transfers/:from_id", handlers.Transfer)
	protected.POST("/deposits/:account_id", handlers.Deposit)
	protected.POST("/accounts/:id/statement", handlers.GetStatement)

	//Loans
	protected.POST("/loans", handlers.CreateLoan)
	protected.GET("/loans", handlers.ListLoans)
	protected.POST("/loans/:id/repay", handlers.MakePayment)
	protected.POST("/loans/:id/payments", handlers.ListPayments)

	//beneficiaries
	protected.POST("/beneficiaries", handlers.AddBeneficiary)

	log.Printf("server starting on %s", port)
	r.Run(":" + port)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "***")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		//handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
