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

func GetHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello from Gin!",
	})
}

func GetHelloWithName(c *gin.Context) {
	name := c.Param("name")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + name + "!",
	})
}

func GetGreet(c *gin.Context) {
	name := c.DefaultQuery("name", "World")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + name + "!",
	})
}

// POST handlers
func PostEcho(c *gin.Context) {
	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data received successfully",
		"data":    jsonData,
	})
}

func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "gin-hello-world",
	})
}
