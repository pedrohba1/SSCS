package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Recorder    RecorderConfig    `yaml:"recorder"`
	Indexer     IndexerConfig     `yaml:"indexer"`
	Recorgnizer RecorgnizerConfig `yaml:"recorgnizer"`
	Cleaner     CleanerConfig     `yaml:"cleaner"`
}

type RecorderConfig struct {
	RTSP          RTSPConfig `yaml:"rtsp"`
	RecordingsDir string     `yaml:"recordingsDir"`
}

type IndexerConfig struct {
	DbUrl string `yaml:"dbUrl"`
}

type RecorgnizerConfig struct {
	FaceHaarPath string `yaml:"faceHaarPath"`
	ThumbsDir    string `yaml:"thumbsDir"`
}

type CleanerConfig struct {
	SizeLimit   int `yaml:"sizeLimit"`
	CheckPeriod int `yaml:"checkPeriod"`
}

type RTSPConfig struct {
	Feeds []string `yaml:"feeds"`
}

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
	return cfg, err
}
