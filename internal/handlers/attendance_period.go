package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/chafid/payroll-project/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendancePeriodRequest struct {
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

func CreateAttendancePeriod(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AttendancePeriodRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		//Parse start and end dates for period
		startDate, err := time.Parse("2006-01-02", req.PeriodStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period start date format"})
			return
		}

		endDate, err := time.Parse("2006-01-02", req.PeriodEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period end date format"})
			return
		}

		if endDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End date cannot be earlier than start date"})
			return
		}
		// Must be same year and month
		if startDate.Year() != endDate.Year() || startDate.Month() != endDate.Month() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start and end dates must be in the same month"})
			return
		}

		// Start must be the 1st of the month
		if startDate.Day() != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date must be the first day of the month"})
			return
		}

		// End must be the last day of the month
		lastDay := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		if !endDate.Equal(lastDay) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be the last day of the same month"})
			return
		}

		// Prevent duplicate period for the same month
		var count int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM attendance_periods 
			WHERE date_trunc('month', start_date) = date_trunc('month', $1::date)
		`, startDate).Scan(&count)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Attendance period for this month already exists"})
			return
		}

		userIDStr, _ := c.Get("user_id")
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id in token"})
			return
		}

		ip := c.ClientIP()
		periodID := fmt.Sprintf("%02d-%d", startDate.Month(), startDate.Year())
		//insert attendance period

		_, err = db.Exec(`
			INSERT INTO attendance_periods (
				id, start_date, end_date,
				created_at, updated_at, created_by, updated_by,
				created_ip, updated_ip			
			) VALUES (
				$1, $2, $3, 
				NOW(), NOW(), $4, $4, $5, $5 
			)
		`, periodID, startDate, endDate, userID, ip)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		changeData, err := json.Marshal(req)
		if err != nil {
			changeData = []byte(`{}`)
		}
		utils.LogAudit(db, "CREATE", "attendance_period", periodID, userID, net.ParseIP(ip), changeData)
	}

}
