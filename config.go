package main

import (
	"io/ioutil"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Path       string                            // Path to configuration file
	Verbose    bool                              // Verbose logging mode?
	StatsdHost string       `yaml:"statsd_host"` // Statsd host to send to
	StatsdPort string       `yaml:"statsd_port"` // Statsd port to send to
	Items      []ConfigItem `yaml:"items"`       // Items to sample
}

type ConfigItem struct {
	Name        string `yaml:"name"`     // Name to send statistic as
	Kind        string `yaml:"type"`     // Type of sample (file, command, etc)
	Interval    int    `yaml:"interval"` // Sampling interval
	Path        string `yaml:"path"`     // Path to file or command to run, etc
	Metric      string `yaml:"metric"`   // Type of metric
	Delta       bool   `yaml:"delta"`    // Delta? (only applies to counter)
	CurrentVal  int64                    // Current sample value
}

func populateConfig(cfg *Config) error {
	data, err := ioutil.ReadFile(cfg.Path)
	if err == nil {
		err = yaml.Unmarshal(data, cfg)
	}
	return err
}
