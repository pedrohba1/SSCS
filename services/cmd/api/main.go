package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/api/controllers"
	"github.com/pedrohba1/SSCS/services/api/models"
)

func main() {
	r := gin.Default()

	models.ConnectDatabase() // new

	r.GET("/recognitions", controllers.FindRecogs)
	r.GET("/recordings", controllers.FindRecordings)
	r.GET("/file/*filepath", controllers.ServeFile)

	r.Run(":3000")
}
