package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"

	"github.com/jawher/mow.cli"
)

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
		} else {
			startSamplers(ctx, cfg)
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
