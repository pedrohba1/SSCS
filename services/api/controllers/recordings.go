package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /recognitions
// Get all recognition events. It is capable to filter by
// dates in unix timestamp
func FindRecordings(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"data": "recogs"})
}
