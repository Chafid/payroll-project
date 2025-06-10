package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/chafid/payroll-project/internal/handlers"
)

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock DB
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Simulate hashed password in DB
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	// Expect DB query and return mock row
	mock.ExpectQuery(`SELECT id, password, role FROM users WHERE username = \$1`).
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "password", "role",
		}).AddRow("1", string(hashedPassword), "admin"))

	// Setup router
	router := gin.Default()
	router.POST("/login", handlers.LoginHandler(db)) // fixed path

	// Create request body
	body := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonBody, _ := json.Marshal(body)

	// Send request
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}
