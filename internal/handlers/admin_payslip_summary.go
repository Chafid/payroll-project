package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPayslipSummaryForAdmin(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		periodID := c.Param("period_id")

		rows, err := db.Query(`
			SELECT u.username, u.id, p.total_take_home
			FROM payslips p
			JOIN users u ON p.user_id = u.id
			WHERE p.attendance_periods_id = $1
			ORDER BY u.username
		`, periodID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payslip summary"})
			return
		}
		defer rows.Close()

		type EmployeeSummary struct {
			Username string  `json:"username"`
			UserID   string  `json:"user_id"`
			TotalPay float64 `json:"total_take_home"`
		}

		var summary []EmployeeSummary
		var grandTotal float64

		for rows.Next() {
			var emp EmployeeSummary
			if err := rows.Scan(&emp.Username, &emp.UserID, &emp.TotalPay); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan result"})
				return
			}
			grandTotal += emp.TotalPay
			summary = append(summary, emp)
		}

		c.JSON(http.StatusOK, gin.H{
			"employees":   summary,
			"grand_total": grandTotal,
		})
	}
}
