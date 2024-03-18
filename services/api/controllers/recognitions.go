package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/api/models"
	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/recognizer"
)

func FindRecogs(c *gin.Context) {
	startDateQuery := c.Query("start_date")
	endDateQuery := c.Query("end_date")

	var recognitions []recognizer.RecognizedEvent
	var err error

	// Create a base query to add conditions dynamically
	query := models.DB.Model(&recognizer.RecognizedEvent{})

	// Check if both start and end dates are provided
	if startDateQuery != "" || endDateQuery != "" {
		startDate, err := time.Parse(time.RFC3339, startDateQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use RFC3339."})
			return
		}
		endDate, err := time.Parse(time.RFC3339, endDateQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use RFC3339."})
			return
		}
		// Ensure endDate includes the whole day by setting the time to the end of the day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}

	err = query.Find(&recognitions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	// appends BaseUrl to the recognitions so it can access via links
	baseUrl := conf.CachedConfig.API.BaseUrl

	for i := range recognitions {
        recognitions[i].Path = baseUrl + "/file/" + recognitions[i].Path
    }

	c.JSON(http.StatusOK, gin.H{"data": recognitions})
}
