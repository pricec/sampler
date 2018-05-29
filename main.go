package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"

	"github.com/go-yaml/yaml"
	"github.com/jawher/mow.cli"
)

type Config struct {
	Path    string                      // Path to configuration file
	Verbose bool                        // Verbose logging mode?
	Items   []ConfigItem `yaml:"items"` // Items to sample
}

type ConfigItem struct {
	Kind     string `yaml:"type"`     // Type of sample (file, command, etc)
	Interval string `yaml:"interval"` // Sampling interval
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

func sigHandler(
	ctx context.Context,
	cancel context.CancelFunc,
	sigChan <- chan os.Signal,
	wantExit *bool,
) {
	select {
	case sig := <- sigChan:
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			*wantExit = true
		}
		cancel()
	case <- ctx.Done():
	}
}

func mainLoop(cfg *Config) {
	wantExit    := false
	sigChan     := make(chan os.Signal)

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for !wantExit {
		ctx, cancel := context.WithCancel(context.Background())
		go sigHandler(context.Background(), cancel, sigChan, &wantExit)
		if err := populateConfig(cfg); err != nil {
			fmt.Printf("Failed to read configuration: %v\n", err)
			sigChan <- syscall.SIGTERM
		}
		<-ctx.Done()
	}
}

func main() {
	app := cli.App("sampler", "Sample values and send to statsd")

	app.Spec = "[-v] CONFIG_FILE"

	var (
		verbose = app.BoolOpt("v verbose", false, "Verbose logging mode")
		cfgFile = app.StringArg("CONFIG_FILE", "", "Path to config file")
	)

	app.Action = func() {
		cfg := Config{
			Path:    *cfgFile,
			Verbose: *verbose,
		}

		mainLoop(&cfg)
	}

	app.Run(os.Args)
}
