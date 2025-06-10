package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chafid/payroll-project/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateAttendancePeriod(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		c.Request.RemoteAddr = "127.0.0.1:1234"
		c.Next()
	})
	router.POST("/period", handlers.CreateAttendancePeriod(db))

	t.Run("Success", func(t *testing.T) {
		start := "2025-06-01"
		end := "2025-06-30"

		// Mock duplicate check
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM attendance_periods`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Mock INSERT
		mock.ExpectExec(`INSERT INTO attendance_periods`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "127.0.0.1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		reqBody := `{"period_start":"` + start + `", "period_end":"` + end + `"}`

		req := httptest.NewRequest(http.MethodPost, "/period", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code) // no explicit success code in handler, so it stays 200
	})

	t.Run("Duplicate Period", func(t *testing.T) {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM attendance_periods`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		reqBody := `{"period_start":"2025-06-01", "period_end":"2025-06-30"}`
		req := httptest.NewRequest(http.MethodPost, "/period", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusConflict, resp.Code)
	})

	t.Run("Invalid Date Format", func(t *testing.T) {
		reqBody := `{"period_start":"bad-date", "period_end":"2025-06-30"}`
		req := httptest.NewRequest(http.MethodPost, "/period", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Different Months", func(t *testing.T) {
		reqBody := `{"period_start":"2025-06-01", "period_end":"2025-07-01"}`
		req := httptest.NewRequest(http.MethodPost, "/period", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Start Not First Day", func(t *testing.T) {
		reqBody := `{"period_start":"2025-06-02", "period_end":"2025-06-30"}`
		req := httptest.NewRequest(http.MethodPost, "/period", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("End Not Last Day", func(t *testing.T) {
		reqBody := `{"period_start":"2025-06-01", "period_end":"2025-06-29"}`
		req := httptest.NewRequest(http.MethodPost, "/period", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
