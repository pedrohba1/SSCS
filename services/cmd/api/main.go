// Provides the primary application entry point for a RESTful API server
// using the Gin web framework. This API server handles various routes associated with
// multimedia recognition and recording services.
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/api/controllers"
	"github.com/pedrohba1/SSCS/services/api/models"
)

// main initializes the Gin router, sets up database connections, and defines various routes
// for the API. It also starts the server on port 3000, listening for incoming requests.
// This function is responsible for configuring routes that manage recognitions, recordings,
// file serving, and serving full recording files in MP4 format.
func main() {
	r := gin.Default()

	models.ConnectDatabase() // new

	r.GET("/recognitions", controllers.FindRecogs)
	r.GET("/recordings", controllers.FindRecordings)
	r.GET("/file/*filepath", controllers.ServeFile)
	r.GET("full-recording", controllers.ServeMp4)
	r.Run(":3000")
}
