package utils

import (
	"time"

	"github.com/chafid/payroll-project/config"
	"github.com/golang-jwt/jwt/v5"
)

// var jwtKey = []byte(os.Getenv("JWT_SECRET"))
//var jwtKey = []byte(config.JwtSecret)

func GenerateJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(72 * time.Hour).Unix(), //token expired in 3 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.JwtSecret))
}
