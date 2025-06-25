package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chafid/payroll-project/config"
	"github.com/chafid/payroll-project/db"

	"github.com/gin-gonic/gin"

	"github.com/chafid/payroll-project/internal/handlers"
	"github.com/chafid/payroll-project/internal/middlewares"
	"github.com/chafid/payroll-project/internal/utils"
)

func main() {
	//Load env variables if present
	config.LoadConfig()

	//Connect to db
	db, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v\n", err)
	}
	defer db.Close()

	r := gin.Default()

	//Public routes
	r.POST("/login", handlers.LoginHandler(db))

	//Routes that needs authentications
	api := r.Group("/api")
	api.Use(middlewares.AuthMiddleware())

	//Admin routes
	adminGroup := api.Group("/admin")
	{
		adminGroup.POST("/attendance-periods", handlers.CreateAttendancePeriod(db))
		adminGroup.POST("/run-payroll", handlers.RunPayroll(db))
		adminGroup.GET("/payroll-summary/:period_id", handlers.GetPayslipSummaryForAdmin(db))
	}

	//employee routes
	employeeGroup := api.Group("/employee")
	{
		employeeGroup.POST("/attendance", handlers.SubmitAttendance(db))
		employeeGroup.POST("/overtime", handlers.SubmitOvertime(db))
		employeeGroup.POST("/reimbursement", handlers.SubmitReimbursement(db))
		employeeGroup.GET("/payslip/:period_id", handlers.GetEmployeePayslip(db))
	}

	port := config.Port

	log.Printf("Starting server on port %s\n", port)
	if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v\n", err)
	}
}
