package main

import (
	"io/ioutil"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Path    string                      // Path to configuration file
	Verbose bool                        // Verbose logging mode?
	Items   []ConfigItem `yaml:"items"` // Items to sample
}

type ConfigItem struct {
	Kind     string `yaml:"type"`     // Type of sample (file, command, etc)
	Interval int    `yaml:"interval"` // Sampling interval
	Path     string `yaml:"path"`     // Path to file or command to run, etc
	Metric   string `yaml:"metric"`   // Type of metric
}

func populateConfig(cfg *Config) error {
	data, err := ioutil.ReadFile(cfg.Path)
	if err == nil {
		err = yaml.Unmarshal(data, cfg)
	}
	return err
}
