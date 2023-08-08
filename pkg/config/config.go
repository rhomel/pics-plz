package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	ImageRoot               string
	ConvertedImageCachePath string
}

func NewConfig() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &Config{
		ConvertedImageCachePath: filepath.Join(wd, "/converted"),
	}, nil
}
