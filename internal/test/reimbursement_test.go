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

func TestSubmitReimbursement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.New()
	router.POST("/reimbursement", func(c *gin.Context) {
		c.Set("user_id", "11111111-1111-1111-1111-111111111111")
		handlers.SubmitReimbursement(db)(c)
	})

	payload := handlers.ReimbursementRequest{
		Amount:      100.50,
		Description: "Taxi reimbursement",
		Date:        time.Now().Format("2006-01-02"),
	}
	body, _ := json.Marshal(payload)

	// Expect INSERT INTO reimbursements with 9 values
	mock.ExpectExec(`INSERT INTO reimbursements`).
		WithArgs(
			sqlmock.AnyArg(), "11111111-1111-1111-1111-111111111111", // id, user_id
			payload.Amount, payload.Description, // amount, description
			sqlmock.AnyArg(),                   // date
			sqlmock.AnyArg(), sqlmock.AnyArg(), // created_by, updated_by
			"127.0.0.1", "127.0.0.1", // created_ip, updated_ip
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/reimbursement", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:1234"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Reimbursement submitted successfully")
}
