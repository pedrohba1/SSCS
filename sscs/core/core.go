package core

import (
	"sscs/conf"
)

type Core struct {
	configPath string
	Config     conf.Config
}

func New(args []string) *Core {
	// Extract the config path from the args or define a default
	// This is just a placeholder; you might extract it differently
	configPath := args[0] // assumes args[0] has the config path

	// Read the configuration using ReadConf
	cfg, err := conf.ReadConf(configPath)
	if err != nil {
		panic(err) // or handle the error more gracefully, based on your application's needs
	}

	// Create a new Core instance with the read configuration
	return &Core{
		configPath: configPath,
		Config:     cfg,
	}
}
