// Package conf defines structures for loading configuration settings from a YAML file.
// These configurations are used to set up different aspects of a multimedia processing system,
// including recording, indexing, recognizing, storing, and API settings.
package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config holds all the configuration segments for the application. Each segment is mapped
// from a corresponding section in the YAML configuration file and is structured according
// to the needs of different components of the system like the recorder, indexer, recognizer,
// storer, and the API configuration.
type Config struct {
	Recorder   RecorderConfig   `yaml:"recorder"`
	Indexer    IndexerConfig    `yaml:"indexer"`
	Recognizer RecognizerConfig `yaml:"recognizer"`
	Storer     StorerConfig     `yaml:"storer"`
	API  APIConfig  `yaml:"api"`
}

// RecorderConfig contains configuration necessary for setting up the recording component,
// including RTSP feed details and the directory for storing recordings.
type RecorderConfig struct {
	RTSP          RTSPConfig `yaml:"rtsp"`
	RecordingsDir string     `yaml:"recordingsDir"`
}

// IndexerConfig specifies the database connection URL for the indexing component
// that handles the storage of metadata about recordings.
type IndexerConfig struct {
	DbUrl string `yaml:"dbUrl"`
}

// RecognizerConfig contains settings for the recognition component, including path
// to Haar cascade files, directories for storing thumbnails, and labels for events and frames
type RecognizerConfig struct {
	HaarPath string `yaml:"haarPath"`
	ThumbsDir    string `yaml:"thumbsDir"`
	EventName string `yaml:"eventName"`
	FrameLabel string `yaml:"frameLabel"`
}

// StorerConfig defines the configuration for the storage manager, which handles
// data retention policies including size limits, check periods for storage management,
// and backup path for data backups.
type StorerConfig struct {
	SizeLimit   int    `yaml:"sizeLimit"`
	CheckPeriod int    `yaml:"checkPeriod"`
	BackupPath  string `yaml:"backupPath"`
}

// RTSPConfig holds the configuration for RTSP feeds, which are used by the recorder
// to capture video streams.
type RTSPConfig struct {
	Feeds []string `yaml:"feeds"`
}

// APIConfig provides the base URL and base path settings for the API server, defining
// how the API is accessed externally.
type APIConfig struct {
	BaseUrl  string `yaml:"baseUrl"`
	BasePath string `yaml:"basePath"`
}

// CachedConfig holds a globally available instance of Config once it is loaded.
// This allows other parts of the application to access configuration details efficiently.
var CachedConfig *Config = nil

// Looks for a config file a a few default locations.
// If it doesn't exist in a location, look for the others,
// until it finds a valid one.
func findConfig() (Config, error) {
	var cfg Config
	// List of potential file paths, in order of precedence
	paths := []string{
		"./sscs.yml",     // current directory
		"$HOME/sscs.yml", // home directory
		"/etc/sscs.yml",  // system-wide configuration
	}

	var lastErr error
	for _, path := range paths {
		// Expand environment variables in the path, if any
		path = os.ExpandEnv(path)
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			// Save the error and try the next path
			lastErr = err
			continue
		}
		// Parse the YAML content
		if err = yaml.Unmarshal(fileBytes, &cfg); err != nil {
			lastErr = err
			continue
		}
		// Successfully read and parsed the configuration
		return cfg, nil
	}

	if lastErr != nil {
		// Return the last encountered error if no valid configuration was found
		return cfg, lastErr
	}

	// This would only be reached if the paths slice is empty
	return cfg, fmt.Errorf("no configuration file paths provided")
}

// ReadConf reads the  github.com/pedrohba1/SSCS/services.yml YAML file and unmarshals it into a Go structure.
func ReadConf() (Config, error) {
	cfg, err := findConfig()
	CachedConfig = &cfg
	return cfg, err
}
