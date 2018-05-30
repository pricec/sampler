package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

type Sample struct {
	name  string
	value int64
}

func startSampleSender(
	ctx context.Context,
	cfg *Config,
) (chan<- Sample, error) {
	var sampleChan chan Sample = nil
	conn, err := net.Dial("udp", cfg.StatsdHost + ":" + cfg.StatsdPort)
	if err == nil {
		sampleChan = make(chan Sample, 100)
		go func() {
			defer close(sampleChan)
			defer conn.Close()

			for {
				select {
				case <- ctx.Done():
					return
				case sample := <-sampleChan:
					// TODO: Discriminate between types (item.Kind)
					//       This assumes the type is a gauge
					if _, err := fmt.Fprintf(
						conn,
						"%v:%v|g\n",
						sample.name,
						sample.value,
					); err != nil {
						fmt.Printf(
							"Error sending '%v:%v' to statsd: %v",
							sample.name, sample.value, err,
						)
					}
				}
			}
		}()
	}
	return sampleChan, err
}

func handleSample(item ConfigItem, data string, sampleChan chan<- Sample) {
	val, err := strconv.ParseInt(strings.Trim(data, "\n"), 10, 64)
	if err != nil {
		fmt.Printf(
			"Failed to convert value '%v' for '%v': %v\n",
			data, item.Name, err,
		)
		return
	}

	//fmt.Printf("Got sample for '%v': %v\n", item.Name, val)
	sampleChan <- Sample{ name: item.Name, value: val }
}

func fileSampler(
	ctx context.Context,
	item ConfigItem,
	sampleChan chan<- Sample,
) {
	for {
		select {
		case <- ctx.Done():
			return
		case <- time.After(time.Second * time.Duration(item.Interval)):
			if data, err := ioutil.ReadFile(item.Path); err != nil {
				fmt.Printf("Error reading '%v': %v\n", item.Path, err)
			} else {
				handleSample(item, string(data), sampleChan)
			}
		}
	}
}

func startSampler(
	ctx context.Context,
	item ConfigItem,
	sampleChan chan<- Sample,
) {
	switch item.Kind {
	case "file":
		go fileSampler(ctx, item, sampleChan)
	default:
		fmt.Printf(
			"Failed to start sampler for '%v'. '%v' is an unrecognized type",
			item.Kind, item.Name,
		)
	}
}

func startSamplers(ctx context.Context, cfg *Config) error {
	sampleChan, err := startSampleSender(ctx, cfg)
	if err == nil {
		for _, item := range cfg.Items {
			startSampler(ctx, item, sampleChan)
		}
	}
	return err
}
