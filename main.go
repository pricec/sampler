package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"
	"github.com/jawher/mow.cli"

	"github.com/pricec/sampler/config"
	"github.com/pricec/sampler/samplers"
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

func mainLoop(cfg *config.Config) {
	wantExit    := false
	sigChan     := make(chan os.Signal)

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for !wantExit {
		ctx, cancel := context.WithCancel(context.Background())
		go sigHandler(context.Background(), cancel, sigChan, &wantExit)
		if err := config.PopulateConfig(cfg); err != nil {
			fmt.Printf("Failed to read configuration: %v\n", err)
			sigChan <- syscall.SIGTERM
		} else {
			var err error
			var sampler samplers.Sampler

			sender, err := samplers.NewSender(
				ctx,
				cfg.Prefix,
				cfg.StatsdHost,
				cfg.StatsdPort,
			)
			if err != nil {
				fmt.Printf("Error start sender: %v\n", err)
				os.Exit(1)
			}

			takers := make([]*samplers.SampleTaker, len(cfg.Items))
			for i, item := range cfg.Items {
				switch (item.Kind) {
				case "file":
					sampler, err = samplers.NewFileSampler(&item)
				case "bash":
					sampler, err = samplers.NewBashSampler(&item)
				case "cpu":
					sampler, err = samplers.NewCpuSampler(&item)
				case "memory":
					sampler, err = samplers.NewMemorySampler(&item)
				default:
					fmt.Printf("Unrecognized sampler type '%v'\n", item.Kind)
					continue
				}

				if err == nil {
					takers[i], err = samplers.NewSampleTaker(
						ctx,
						&item,
						sender,
						sampler,
					)
				}

				if err != nil {
					fmt.Printf(
						"Failed to start sample taker for %v: %v\n",
						item.Name,
						err,
					)
					os.Exit(1)
				}
			}
			<-ctx.Done()
		}
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
		cfg := config.Config{
			Path:    *cfgFile,
			Verbose: *verbose,
		}

		mainLoop(&cfg)
	}

	app.Run(os.Args)
}
