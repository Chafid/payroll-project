package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/chafid/payroll-project/internal/models"
	"github.com/chafid/payroll-project/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetEmployeePayslip(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(string)
		periodID := c.Param("period_id")

		var payslip models.Payslip

		err := db.QueryRow(`
			SELECT p.id, p.user_id, u.username, p.attendance_periods_id, p.base_salary, p.attendance_amount,
				p.attendance_days, p.overtime_hours, p.overtime_amount, p.reimbursement_amount, p.total_take_home,
				p.created_at
			FROM payslips p
			JOIN users u ON p.user_id = u.id
			WHERE p.user_id = $1 AND p.attendance_periods_id = $2
		`, userID, periodID).Scan(
			&payslip.ID, &payslip.UserID, &payslip.Username, &payslip.AttendancePeriodID,
			&payslip.BaseSalary, &payslip.AttendanceAmount, &payslip.AttendanceDays,
			&payslip.OvertimeHours, &payslip.OvertimeAmount, &payslip.ReimbursementAmount,
			&payslip.TotalTakeHome, &payslip.CreatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Payslip not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var baseSalary float64
		err = db.QueryRow(`
			SELECT l.base_salary FROM employee_levels l
			JOIN users u ON u.level_id = l.id WHERE u.id = $1
		`, userID).Scan(&baseSalary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch level salary"})
			return
		}

		// Step 2.1: Fetch attendance period date range
		var periodStart, periodEnd time.Time
		err = db.QueryRow(`
			SELECT start_date, end_date FROM attendance_periods WHERE id = $1
			`, periodID).Scan(&periodStart, &periodEnd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance period range"})
			return
		}

		workingDays := utils.CountWorkingDays(periodStart, periodEnd)

		hourlyRate := baseSalary / (float64(workingDays) * 8)
		overtimeRate := hourlyRate * 2

		reimbursements := []models.Reimbursement{}
		rows, err := db.Query(`
			SELECT id, date, description, amount, created_at
			FROM reimbursements
			WHERE user_id = $1 AND date BETWEEN $2 AND $3
		`, userID, periodStart, periodEnd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reimbursements - " + err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var r models.Reimbursement
			err := rows.Scan(&r.ID, &r.Date, &r.Description, &r.Amount, &r.SubmittedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan reimbursement"})
				return
			}
			reimbursements = append(reimbursements, r)
		}

		response := models.PayslipDetailResponse{
			Payslip: payslip,
			Attendance: models.AttendanceBreakdown{
				WorkingDays:      workingDays,
				AttendanceDays:   payslip.AttendanceDays,
				AttendanceAmount: payslip.AttendanceAmount,
			},
			Overtime: models.OvertimeBreakdown{
				OvertimeHours:  nullToZero(payslip.OvertimeHours),
				HourlyRate:     hourlyRate,
				OvertimeRate:   overtimeRate,
				OvertimeAmount: nullToZero(payslip.OvertimeAmount),
			},
			Reimbursements: reimbursements,
		}

		c.JSON(http.StatusOK, response)
	}
}

func nullToZero(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}
