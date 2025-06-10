package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/chafid/payroll-project/internal/utils"
	"github.com/gin-gonic/gin"
)

func RunPayroll(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PeriodID string `json:"period_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get the attendance period
		var startDate, endDate time.Time
		err := db.QueryRow(`
			SELECT start_date, end_date FROM attendance_periods WHERE id = $1
		`, req.PeriodID).Scan(&startDate, &endDate)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Attendance period not found"})
			return
		}

		// Calculate working days
		workingDays := utils.CountWorkingDays(startDate, endDate)
		ip := c.ClientIP()
		userID, exist := c.Get("user_id")
		if !exist {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		query := `
			INSERT INTO payslips (
				user_id, attendance_periods_id, base_salary, attendance_amount, attendance_days,
				overtime_amount, overtime_hours, reimbursement_amount, total_take_home, created_at, created_by, created_ip
			)
				SELECT 
					u.id AS user_id,
					ap.id AS period_id,
					l.base_salary AS base_salary,
					COALESCE(a.attendance_amount, 0) AS attendance_amount,
					COALESCE(a.attendance_days, 0) AS attendance_days,
					COALESCE(o.overtime_amount, 0) AS overtime_amount,
					COALESCE(o.overtime_hours, 0) AS overtime_hours,
					COALESCE(r.reimbursement_amount, 0) AS reimbursement_amount,
					COALESCE(a.attendance_amount, 0) +
					COALESCE(o.overtime_amount, 0) +
					COALESCE(r.reimbursement_amount, 0) AS total_take_home,
					NOW() AS created_at,
					$3 AS created_by,
					$4 AS created_ip
				FROM users u
				JOIN employee_levels l ON u.level_id = l.id
				JOIN attendance_periods ap ON ap.id = $1

				-- Pre-aggregated attendance
				LEFT JOIN (
					SELECT 
						user_id,
						COUNT(*) AS attendance_days,
						(COUNT(*) * 8) * (l.base_salary / ($2 * 8)) AS attendance_amount
					FROM attendances a
					JOIN employee_levels l ON l.id = (
						SELECT level_id FROM users WHERE id = a.user_id
					)
					WHERE period_id = $1
					GROUP BY user_id, l.base_salary
				) a ON u.id = a.user_id

				-- Pre-aggregated overtime
				LEFT JOIN (
					SELECT 
						user_id,
						SUM(hours) AS overtime_hours,
						SUM(hours) * (l.base_salary / ($2 * 8)) * 2 AS overtime_amount
					FROM overtimes o
					JOIN employee_levels l ON l.id = (
						SELECT level_id FROM users WHERE id = o.user_id
					)
					WHERE date BETWEEN (
						SELECT start_date FROM attendance_periods WHERE id = $1
					) AND (
						SELECT end_date FROM attendance_periods WHERE id = $1
					)
					GROUP BY user_id, l.base_salary
				) o ON u.id = o.user_id

				-- Pre-aggregated reimbursements
				LEFT JOIN (
					SELECT 
						user_id,
						SUM(amount) AS reimbursement_amount
					FROM reimbursements
					WHERE date BETWEEN (
						SELECT start_date FROM attendance_periods WHERE id = $1
					) AND (
						SELECT end_date FROM attendance_periods WHERE id = $1
					)
					GROUP BY user_id
				) r ON u.id = r.user_id

				-- Filter only employees with data
				WHERE 
					COALESCE(a.attendance_days, 0) > 0 OR
					COALESCE(o.overtime_hours, 0) > 0 OR
					COALESCE(r.reimbursement_amount, 0) > 0

				ORDER BY u.id;
		`

		_, err = db.Exec(query, req.PeriodID, workingDays, userID, ip)
		if err != nil {
			log.Printf("[RunPayroll] Failed: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payroll"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payroll processed successfully"})
	}
}
