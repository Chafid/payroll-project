package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chafid/payroll-project/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetPayslipSummaryForAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.Default()
	router.GET("/admin/payslip-summary/:period_id", handlers.GetPayslipSummaryForAdmin(db))

	// Mock rows that match: username, user_id, total_pay
	mockRows := sqlmock.NewRows([]string{"username", "id", "total_take_home"}).
		AddRow("employee001", "user1", 3000.0).
		AddRow("employee002", "user2", 2500.0)

	mock.ExpectQuery(`SELECT u.username, u.id, p.total_take_home FROM payslips p`).
		WithArgs("06-2025").
		WillReturnRows(mockRows)

	req := httptest.NewRequest(http.MethodGet, "/admin/payslip-summary/06-2025", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"username":"employee001"`)
	assert.Contains(t, w.Body.String(), `"username":"employee002"`)
	assert.Contains(t, w.Body.String(), `"grand_total":5500`)
}
