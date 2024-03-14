package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/cmd/api/controllers"
	"github.com/pedrohba1/SSCS/services/cmd/api/models"
)

func main() {
	r := gin.Default()

	
	models.ConnectDatabase() // new

	r.GET("/recognitions", controllers.FindRecogs)
	r.GET("/recordings", controllers.FindRecordings)

	// Start the server
	r.Run(":8080")
}
