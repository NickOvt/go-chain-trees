package api

import "github.com/gin-gonic/gin"

// SetupRoutes configures all the routes for the API
func SetupRoutes(r *gin.Engine) {
	// Hello World endpoints
	r.GET("/", GetRoot)

	// Health check endpoint
	r.GET("/health", GetHealth)
}
