package core

import (
	"sscs/conf"
	BaseLogger "sscs/logger"

	"github.com/sirupsen/logrus"
)

type Core struct {
	configPath string
	config     conf.Config
	Logger     *logrus.Entry
}

func New(args []string) *Core {
	// Extract the config path from the args or define a default
	// This is just a placeholder; you might extract it differently
	configPath := args[0] // assumes args[0] has the config path

	// Read the configuration using ReadConf
	cfg, err := conf.ReadConf("./sscs.yml")
	if err != nil {
		panic(err) // or handle the error more gracefully, based on your application's needs
	}

	// Create a new Core instance with the read configuration

	return &Core{
		configPath: configPath,
		config:     cfg,
		Logger:     BaseLogger.BaseLogger.WithField("package", "core"),
	}
}
