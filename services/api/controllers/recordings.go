package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /recordings
// Gets all recordings between two dates as a single served file.
// dates have to be passed in Unix timestamp
func FindRecordings(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"data": "recogs"})
}
