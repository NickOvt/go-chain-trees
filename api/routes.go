package api

import "github.com/gin-gonic/gin"

// SetupRoutes configures all the routes for the API
func SetupRoutes(r *gin.Engine) {
	// Hello World endpoints
	r.GET("/", GetRoot)
	r.GET("/hello", GetHello)
	r.GET("/hello/:name", GetHelloWithName)
	r.GET("/greet", GetGreet)

	// POST endpoints
	r.POST("/echo", PostEcho)

	//// User CRUD endpoints
	//r.POST("/users", CreateUser)
	//r.GET("/users", GetUsers)
	//r.GET("/users/:id", GetUserByID)
	//r.PUT("/users/:id", UpdateUser)
	//r.DELETE("/users/:id", DeleteUser)

	// Health check endpoint
	r.GET("/health", GetHealth)
}
