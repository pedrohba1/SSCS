package conf

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	RTSP RTSPConfig `yaml:"rtsp"`
}

type RTSPConfig struct {
	Feeds []string `yaml:"feeds"`
}

// ReadConf reads a YAML file and unmarshals it into a Go structure.
func ReadConf(filename string) (Config, error) {
	// Read the file
	var cfg Config
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return cfg, err
	}

	// Parse the YAML content
	err = yaml.Unmarshal(fileBytes, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
