package testutils

import (
	"github.com/gin-gonic/gin"
)

// Mock middleware to inject employee user
func InjectMockUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", "1") // use realistic test user ID
		c.Set("isAdmin", false)
		c.Next()
	}
}

// Mock middleware to inject admin user
func InjectMockAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", "1")
		c.Set("isAdmin", true)
		c.Next()
	}
}
