package config

import (
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	WorkingDir string
	ConfFile   string
	ConfDir    string
}

func InitConfig() *Config {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get working directory")
	}
	return &Config{
		WorkingDir: workingDir,
		ConfDir:    filepath.Join(workingDir, ".vulta"),
		ConfFile:   filepath.Join(workingDir, ".vulta", "vulta.yaml"),
	}
}
