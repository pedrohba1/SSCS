package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/api/models"
	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/recorder"
)

// GET /recordings
// Gets all recordings files between two dates
// dates have to be passed in Unix timestamp
func FindRecordings(c *gin.Context) {
	startDateQuery := c.Query("start_date")
	endDateQuery := c.Query("end_date")

	var recordings []recorder.RecordedEvent
	
	query := models.DB.Model(&recorder.RecordedEvent{})

	
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

	err := query.Find(&recordings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	baseUrl := conf.CachedConfig.API.BaseUrl


	for i := range recordings {
		baseIndex := strings.Index(recordings[i].Path, "recordings")
		if baseIndex == -1 {
			fmt.Println("Base directory not found in the path")
			continue
		}

        recordings[i].Path = baseUrl + "/file/" + recordings[i].Path[baseIndex:]
    }

	c.JSON(http.StatusOK, gin.H{"data": recordings})
}
