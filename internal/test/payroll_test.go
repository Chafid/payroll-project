package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chafid/payroll-project/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRunPayroll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.Default()

	// Inject mock user_id middleware
	router.POST("/run-payroll", func(c *gin.Context) {
		c.Set("user_id", "admin-user") // simulate admin user
		handlers.RunPayroll(db)(c)
	})

	// Step 1: Mock attendance_periods lookup
	mock.ExpectQuery(`SELECT start_date, end_date FROM attendance_periods WHERE id = \$1`).
		WithArgs("06-2025").
		WillReturnRows(sqlmock.NewRows([]string{"start_date", "end_date"}).
			AddRow(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)))

	// Step 2: Mock INSERT INTO payslips
	mock.ExpectExec(`INSERT INTO payslips`).
		WithArgs("06-2025", sqlmock.AnyArg(), "admin-user", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1)) // pretend one row inserted

	// Step 3: Prepare request
	payload := map[string]string{
		"period_id": "06-2025",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/run-payroll", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Payroll processed successfully")
}
