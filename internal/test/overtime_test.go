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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSubmitOvertime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.New()
	router.POST("/overtime", func(c *gin.Context) {
		c.Set("user_id", "11111111-1111-1111-1111-111111111111")
		c.Request = c.Request.WithContext(c)
		handlers.SubmitOvertime(db)(c)
	})

	payload := handlers.OvertimeRequest{
		Date:  time.Now().Format("2006-01-02"),
		Hours: 2,
	}
	body, _ := json.Marshal(payload)

	// Mock insert or update query
	mock.ExpectExec(`INSERT INTO overtimes`).
		WithArgs(sqlmock.AnyArg(), // overtime ID
			uuid.MustParse("11111111-1111-1111-1111-111111111111"), // user ID
			sqlmock.AnyArg(), // date
			payload.Hours,    // hours
			uuid.MustParse("11111111-1111-1111-1111-111111111111"), // created_by
			"127.0.0.1", // IP
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Optional: you can skip LogAudit or mock it if needed

	req := httptest.NewRequest(http.MethodPost, "/overtime", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:1234"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
