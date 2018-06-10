package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

type Sample struct {
	name   string
	value  int64
	metric string
}

func sendSample(
	conn *net.Conn,
	kind string,
	name string,
	value int64,
) {
	var extension string
	switch (kind) {
	case "counter":
		extension = "c"
	case "set":
		extension = "s"
	case "gauge":
		extension = "g"
	default:
		fmt.Printf("Unrecognized metric type %v\n", kind)
		return
	}

	stat := fmt.Sprintf("%v:%v|%v\n", name, value, extension)

	if _, err := fmt.Fprintf(*conn, stat); err != nil {
		fmt.Printf("Error sending '%v:%v' to statsd: %v", name, value, err)
	}
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
					sendSample(
						&conn,
						sample.metric,
						cfg.Prefix + "." + sample.name,
						sample.value,
					)
				}
			}
		}()
	}
	return sampleChan, err
}

func handleSample(item *ConfigItem, data string, sampleChan chan<- Sample) {
	val, err := strconv.ParseInt(strings.Trim(data, "\n"), 10, 64)
	if err != nil {
		fmt.Printf(
			"Failed to convert value '%v' for '%v': %v\n",
			data, item.Name, err,
		)
		return
	}

	// Adjust sample value if this is a delta counter
	if item.Metric == "counter" && item.Delta {
		// Don't send an update at all if this is the first sample
		// (for delta, we need a history for the sample to make sense)
		if !item.Initialized {
			item.Initialized = true
			item.CurrentVal = val
			return
		} else {
			val -= item.CurrentVal
			item.CurrentVal += val
		}
	} else {
		item.CurrentVal = val
	}

	//fmt.Printf("Got sample for '%v': %v\n",	item.Name, val)
	sampleChan <- Sample{ name: item.Name, value: val, metric: item.Metric }
}

func bashSampler(item *ConfigItem) (string, error) {
	output, err := exec.Command("/bin/bash", "-c", item.Path).Output()
	return string(output), err
}

func fileSampler(item *ConfigItem) (string, error) {
	data, err := ioutil.ReadFile(item.Path)
	return string(data), err
}

func sampler(
	ctx context.Context,
	item *ConfigItem,
	sampleChan chan<- Sample,
	sampleFunc func(*ConfigItem) (string, error),
) {
	for {
		select {
		case <- ctx.Done():
			return
		case <- time.After(time.Second * time.Duration(item.Interval)):
			if sample, err := sampleFunc(item); err != nil {
				fmt.Printf("Error reading '%v': %v\n", item.Path, err)
			} else {
				handleSample(item, sample, sampleChan)
			}
		}
	}
}

func startSampler(
	ctx context.Context,
	item ConfigItem,
	sampleChan chan<- Sample,
) {
	var sampleFunc func(*ConfigItem) (string, error)
	switch item.Kind {
	case "file":
		sampleFunc = fileSampler
	case "bash":
		sampleFunc = bashSampler
	default:
		fmt.Printf(
			"Failed to start sampler for '%v'. '%v' is an unrecognized type",
			item.Kind, item.Name,
		)
		return
	}

	go sampler(ctx, &item, sampleChan, sampleFunc)
}

func startSamplers(ctx context.Context, cfg *Config) error {
	sampleChan, err := startSampleSender(ctx, cfg)
	if err == nil {
		for _, item := range cfg.Items {
			// TODO: I want to pass in a pointer to startSampler, but
			//       for some reason which I do not yet understand,
			//       it can be changed after the fact. That is, every
			//       sampler will sample the same item if I replace
			//       'item' with '&item' below.
			startSampler(ctx, item, sampleChan)
		}
	}
	return err
}
