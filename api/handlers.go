package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Hello World handlers
func GetRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
		"service": "Gin Hello World API",
	})
}

func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "gin-hello-world",
	})
}
