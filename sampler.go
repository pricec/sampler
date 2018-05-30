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
					sendSample(&conn, sample.metric, sample.name, sample.value)
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
		val -= item.CurrentVal
		item.CurrentVal += val
	} else {
		item.CurrentVal = val
	}

	//fmt.Printf("Got sample for '%v': %v\n",	item.Name, val)
	sampleChan <- Sample{ name: item.Name, value: val, metric: item.Metric }
}

func fileSampler(
	ctx context.Context,
	item *ConfigItem,
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
	item *ConfigItem,
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
			startSampler(ctx, &item, sampleChan)
		}
	}
	return err
}
