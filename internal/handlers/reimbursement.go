package handlers

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/chafid/payroll-project/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReimbursementRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
}

func SubmitReimbursement(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReimbursementRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDStr := c.GetString("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id in token"})
			return
		}

		ip := c.ClientIP()

		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
			return
		}

		id := uuid.New()

		_, err = db.Exec(`
			INSERT INTO reimbursements 
			(id, user_id, amount, description, date, created_by, updated_by, created_ip, updated_ip)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, id, userID, req.Amount, req.Description, parsedDate, userID, userID, ip, ip)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit reimbursement"})
			return
		}

		//Audit log
		changeData := map[string]interface{}{
			"user_id":     userID,
			"amount":      req.Amount,
			"description": req.Description,
			"date":        req.Date,
		}

		jsonBytes, _ := json.Marshal(changeData)
		utils.LogAudit(db, "INSERT", "reimbursements", id.String(), userID, net.ParseIP(ip), jsonBytes)

		c.JSON(http.StatusOK, gin.H{"message": "Reimbursement submitted successfully"})
	}
}
