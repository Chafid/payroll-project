package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chafid/payroll-project/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetEmployeePayslip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.Default()
	router.GET("/payslip/:period_id", func(c *gin.Context) {
		c.Set("user_id", "11111111-1111-1111-1111-111111111111")
		handlers.GetEmployeePayslip(db)(c)
	})

	// 1. Mock payslip
	mock.ExpectQuery(`SELECT p\.id, p\.user_id, u\.username, p\.attendance_periods_id, p\.base_salary, p\.attendance_amount, p\.attendance_days, p\.overtime_hours, p\.overtime_amount, p\.reimbursement_amount, p\.total_take_home, p\.created_at FROM payslips p JOIN users u ON p\.user_id = u\.id WHERE p\.user_id = \$1 AND p\.attendance_periods_id = \$2`).
		WithArgs("11111111-1111-1111-1111-111111111111", "06-2025").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "username", "attendance_periods_id", "base_salary",
			"attendance_amount", "attendance_days", "overtime_hours",
			"overtime_amount", "reimbursement_amount", "total_take_home", "created_at",
		}).AddRow(
			"p1", "11111111-1111-1111-1111-111111111111", "employee123", "06-2025", 3000.0,
			2700.0, 20, 10.0, 300.0, 50.0, 3050.0, time.Now(),
		))

	// 2. Mock level salary (updated to match JOIN with users)
	mock.ExpectQuery(`SELECT l\.base_salary FROM employee_levels l JOIN users u ON u\.level_id = l\.id WHERE u\.id = \$1`).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnRows(sqlmock.NewRows([]string{"salary"}).AddRow(3000.0))

	// 3. Mock attendance period dates
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT start_date, end_date FROM attendance_periods`).
		WithArgs("06-2025").
		WillReturnRows(sqlmock.NewRows([]string{"start_date", "end_date"}).
			AddRow(start, end))

		// 4. Mock reimbursements
	mock.ExpectQuery(`SELECT id, date, description, amount, created_at FROM reimbursements WHERE user_id = \$1 AND date BETWEEN \$2 AND \$3`).
		WithArgs("11111111-1111-1111-1111-111111111111", start, end).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "date", "description", "amount", "created_at",
		}).AddRow("r1", start.AddDate(0, 0, 5), "Internet", 50.0, time.Now()))

	// Perform request
	req := httptest.NewRequest("GET", "/payslip/06-2025", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"total_take_home":3050`)
	assert.Contains(t, w.Body.String(), `"overtime_hours":10`)
	assert.Contains(t, w.Body.String(), `"reimbursements"`)
	assert.Contains(t, w.Body.String(), `"username":"employee123"`)

}
