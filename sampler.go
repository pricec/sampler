package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

func handleSample(item ConfigItem, data string) {
	val, err := strconv.ParseInt(strings.Trim(data, "\n"), 10, 64)
	if err != nil {
		fmt.Printf(
			"Failed to convert value '%v' for '%v': %v\n",
			data, item.Path, err,
		)
		return
	}

	// TODO: Send the value to statsd
	fmt.Printf("Got sample for '%v': %v\n", item.Path, val)
}

func fileSampler(ctx context.Context, item ConfigItem) {
	for {
		select {
		case <- ctx.Done():
			return
		case <- time.After(time.Second * time.Duration(item.Interval)):
			if data, err := ioutil.ReadFile(item.Path); err != nil {
				fmt.Printf("Error reading '%v': %v\n", item.Path, err)
			} else {
				handleSample(item, string(data))
			}
		}
	}
}

func startSampler(ctx context.Context, item ConfigItem) {
	switch item.Kind {
	case "file":
		go fileSampler(ctx, item)
	default:
		fmt.Printf(
			"Failed to start sampler for '%v'. '%v' is an unrecognized type",
			item.Kind, item.Path,
		)
	}
}

func startSamplers(ctx context.Context, cfg *Config) {
	for _, item := range cfg.Items {
		startSampler(ctx, item)
	}
}
