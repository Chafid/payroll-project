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

// Request payload
type AttendanceRequest struct {
	PeriodID string `json:"period_id" binding:"required"`
	Date     string `json:"date" binding:"required"` //format YYYY-MM-DD
}

func SubmitAttendance(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AttendanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		//parse date string
		attendanceDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		//check for weekend
		if attendanceDate.Weekday() == time.Saturday || attendanceDate.Weekday() == time.Sunday {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot submit attendance on weekends"})
			return
		}

		//get user info from middleware
		userIDStr, exist := c.Get("user_id")
		if !exist {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id in token"})
			return
		}

		//validate the given period exist and date is within the period
		var startDate, endDate time.Time
		err = db.QueryRow(`SELECT start_date, end_date from attendance_periods WHERE id = $1`, req.PeriodID).Scan(&startDate, &endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period_id"})
			return
		}
		if attendanceDate.Before(startDate) || attendanceDate.After(endDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Attendance date is not within the attendance period"})
			return
		}

		ip := c.ClientIP()
		attendanceID := uuid.New()
		//Insert the attendance
		_, err = db.Exec(`
			INSERT INTO attendances (id, user_id, date, period_id, created_by, updated_by, created_ip, updated_ip )
			VALUES ($1, $2, $3, $4, $5, $5, $6, $6)
		  `, attendanceID, userID, attendanceDate, req.PeriodID, userID, c.ClientIP())

		if err != nil {
			if utils.IsUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "Attendance already submitted for this date"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}

			return
		}

		//audit log
		changeData, err := json.Marshal(req)
		if err != nil {
			changeData = []byte(`{}`)
		}
		utils.LogAudit(db, "INSERT", "attendance", attendanceID.String(), userID, net.ParseIP(ip), changeData)

		c.JSON(http.StatusCreated, gin.H{"message": "Attendance is successfully submitted"})

	}
}
