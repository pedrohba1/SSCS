package controllers

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/conf"
)

// ServeFile dynamically serves files based on the provided URL path
// can be used to either fetch individual recordings or recognition
// images
func ServeFile(c *gin.Context)  {
		// Extract the filepath from the URL
		filepath := c.Param("filepath")

		// Use the base path from the config
		basePath := conf.CachedConfig.API.BasePath

		// Combine the base path with the requested file path
		fullPath := path.Join(basePath, filepath)

		// Serve the file. Ensure the path is sanitized and safe to use.
		c.File(fullPath)
	
}