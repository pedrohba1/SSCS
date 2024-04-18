package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pedrohba1/SSCS/services/api/models"
	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/recorder"
)

// AggregateTSFilesToMP4 takes a slice of .ts file paths and an output .mp4 file path, combines the .ts files, and converts the result to .mp4.
func AggregateTSFilesToMP4(tsFiles []string, outputDir, outputFile string) error {
    // Create a temporary file to list .ts files
    listFile, err := filepath.Abs("filelist.txt")
    if err != nil {
        return err
    }
    defer os.Remove(listFile) // Clean up after

    // Write file paths to the list file
    file, err := os.Create(listFile)
    if err != nil {
        return err
    }

    for _, tsFile := range tsFiles {
        _, err := file.WriteString("file '" + tsFile + "'\n")
        if err != nil {
            file.Close()
            return err
        }
    }
    file.Close()

    // Construct the full output file path
    fullOutputPath := filepath.Join(outputDir, outputFile)

    // Execute FFmpeg command to concatenate and convert .ts to .mp4
    // Adding -y to overwrite existing files without asking
    cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", fullOutputPath)
    if err := cmd.Run(); err != nil {
        fmt.Println("ERROR: ", err)
        return err
    }

    // Optionally, print the full path of the created file
    fmt.Println("File created at:", fullOutputPath)

    return nil
}


// ServeFile dynamically serves files based on the provided URL path
// can be used to either fetch individual recordings or recognition
// images
func ServeMp4(c *gin.Context)  {
	startDateQuery := c.Query("start_date")
	endDateQuery := c.Query("end_date")
	// Extract the filepath from the URL

	// Use the base path from the config
	// basePath := conf.CachedConfig.API.BasePath
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
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		query = query.Where("start_time BETWEEN ? AND ?", startDate, endDate)
	}

	err := query.Find(&recordings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var validPaths []string

	for i := range recordings {	
		fmt.Println(recordings[i].Path)

		absPath, err := filepath.Abs(recordings[i].Path)
		if err != nil {
			log.Printf("Error resolving path: %s, Error: %s\n", recordings[i].Path, err)
			continue
		}

		// Check if the file exists
		if _, err := os.Stat(absPath); err == nil {
			validPaths = append(validPaths, absPath)
		} else {
			log.Printf("File does not exist: %s\n", absPath)
		}
    }
	startTime := recordings[0].StartTime;
	endTime:= recordings[len(recordings) -1].EndTime;

	// Convert times to string and replace colons and spaces
	startTimeString := strings.ReplaceAll(startTime.Format(time.RFC3339), ":", "-")
	startTimeString = strings.ReplaceAll(startTimeString, " ", "") // Remove spaces

	endTimeString := strings.ReplaceAll(endTime.Format(time.RFC3339), ":", "-")
	endTimeString = strings.ReplaceAll(endTimeString, " ", "") // Remove spaces

	// Combine strings into a safe filename
	safeFilename := fmt.Sprintf("%s-%s.mp4", startTimeString, endTimeString)
		
	outputDir := conf.CachedConfig.Recorder.RecordingsDir
	// Serve the file. Ensure the path is sanitized and safe to use.

	err = AggregateTSFilesToMP4(validPaths,outputDir,safeFilename);
	

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


    fullOutputPath := filepath.Join(outputDir, safeFilename)
	baseIndex := strings.Index(fullOutputPath, "recordings")
	if baseIndex == -1 {
		fmt.Println("Base directory not found in the path")
	}

	baseUrl := conf.CachedConfig.API.BaseUrl

	hyperlink := baseUrl + "/file/" + fullOutputPath[baseIndex:]

	c.JSON(http.StatusOK, gin.H{"data": hyperlink})

}