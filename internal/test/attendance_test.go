package test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chafid/payroll-project/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSubmitAttendance(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := handlers.SubmitAttendance(db)

	// Define valid test data
	userID := uuid.New()
	periodID := "06-2025"
	attendanceDate := time.Now().AddDate(0, 0, 1) // Tomorrow (make sure it's not weekend)

	startDate := attendanceDate.AddDate(0, 0, -5)
	endDate := attendanceDate.AddDate(0, 0, 5)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT start_date, end_date from attendance_periods WHERE id = \$1`).
			WithArgs(periodID).
			WillReturnRows(sqlmock.NewRows([]string{"start_date", "end_date"}).
				AddRow(startDate, endDate))

		mock.ExpectExec(`INSERT INTO attendances`).
			WithArgs(sqlmock.AnyArg(), userID, sqlmock.AnyArg(), periodID, userID, "127.0.0.1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Create request
		payload := `{"period_id": "` + periodID + `", "date": "` + attendanceDate.Format("2006-01-02") + `"}`

		req := httptest.NewRequest(http.MethodPost, "/attendance", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Gin context with user ID and IP
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID.String())
		ctx.Request.RemoteAddr = "127.0.0.1:1234"

		handler(ctx)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "successfully")
	})

	t.Run("Weekend error", func(t *testing.T) {
		saturday := time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC) // Ensure weekend

		payload := `{"period_id": "` + periodID + `", "date": "` + saturday.Format("2006-01-02") + `"}`

		req := httptest.NewRequest(http.MethodPost, "/attendance", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)

		handler(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "weekends")
	})

	t.Run("Invalid period_id", func(t *testing.T) {
		mock.ExpectQuery(`SELECT start_date, end_date from attendance_periods WHERE id = \$1`).
			WithArgs(periodID).
			WillReturnError(sql.ErrNoRows)

		payload := `{"period_id": "` + periodID + `", "date": "` + attendanceDate.Format("2006-01-02") + `"}`

		req := httptest.NewRequest(http.MethodPost, "/attendance", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID.String())

		handler(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid period_id")
	})
}
