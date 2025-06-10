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

type OvertimeRequest struct {
	Date  string  `json:"date" binding:"required"` //YYYY-MM-DD
	Hours float64 `json:"hours" binding:"required"`
}

func SubmitOvertime(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OvertimeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		//validate hours
		if req.Hours <= 0 || req.Hours > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Overtime must between 1 to 3 hours"})
			return
		}

		overtimeDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong overtime date format"})
			return
		}

		userIDStr := c.GetString("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id in token"})
			return
		}

		ip := c.ClientIP()
		overtimeID := uuid.New()

		_, err = db.Exec(`
			INSERT INTO overtimes (id, user_id, date, hours, created_by, created_ip) 
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, date) DO UPDATE 
			SET hours = EXCLUDED.hours, 
				updated_at = now(),
				updated_by = $5,
				updated_ip = $6
			`, uuid.New(), userID, overtimeDate, req.Hours, userID, ip)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//insert into audit logs
		changeData, err := json.Marshal(req)
		if err != nil {
			changeData = []byte(`{}`)
		}
		utils.LogAudit(db, "INSERT/UPDATE", "overtimes", overtimeID.String(), userID, net.ParseIP(ip), changeData)
	}
}
