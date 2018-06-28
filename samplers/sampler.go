package samplers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/pricec/sampler/config"
)

type MetricType int

const (
	METRIC_TYPE_COUNTER = iota
	METRIC_TYPE_SET
	METRIC_TYPE_GAUGE
)

var StringToMetricType = map[string]MetricType{
	"counter": METRIC_TYPE_COUNTER,
	"set"    : METRIC_TYPE_SET,
	"gauge"  : METRIC_TYPE_GAUGE,
}

type Sample struct {
	name   string
	value  int64
	metric MetricType
	suffix string
}

type SampleTaker struct {
	name        string
	sender      *Sender
	interval    int
	metric      MetricType
	delta       bool
	valMap      map[string]int64
	sampler     Sampler
}

func NewSampleTaker(
	ctx context.Context,
	item *config.ConfigItem,
	sender *Sender,
	sampler Sampler,
) (*SampleTaker, error) {
	metric, ok := StringToMetricType[item.Metric]
	if !ok {
		return nil, errors.New(
			fmt.Sprintf("Unknown metric type '%v'\n", item.Metric),
		)
	}

	taker := &SampleTaker{
		name: item.Name,
		sender: sender,
		interval: item.Interval,
		metric: metric,
		delta: item.Delta,
		valMap: map[string]int64{},
		sampler: sampler,
	}

	return taker, taker.start(ctx)
}

func (s *SampleTaker) start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			case <- time.After(time.Duration(s.interval) * time.Second):
				if valMap, err := s.sampler.Sample(); err != nil {
					fmt.Printf("Error sampling '%v': %v\n", s.name, err)
				} else {
					for field, val := range valMap {
						if val, skip := s.adjust(field, val); !skip {
							s.sender.Send(
								Sample{ s.name, val, s.metric , field},
							)
						}
					}
				}
			}
		}
	}()
	return nil
}

// Return done = true if the item is uninitialized. Also updates
// the most recent value (current) and the initialized flag.
func (s *SampleTaker) adjust(field string, inVal int64) (int64, bool) {
	defer func() { s.valMap[field] = inVal }()

	if s.delta {
		current, ok := s.valMap[field]
		if ok {
			return inVal - current, false
		} else {
			return inVal, true
		}
	} else {
		return inVal, false
	}
}

type Sampler interface {
	Sample() (map[string]int64, error)
}

type FileSampler struct {
	path string
}

func NewFileSampler(item *config.ConfigItem) (*FileSampler, error) {
	return &FileSampler{ path: item.Path }, nil
}

func (s *FileSampler) Sample() (map[string]int64, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	val, err := strconv.ParseInt(strings.Trim(string(data), "\n"), 10, 64)
	if err != nil {
		return nil, err
	}

	return map[string]int64{ "": val }, nil
}

type BashSampler struct {
	command string
}

func NewBashSampler(item *config.ConfigItem) (*BashSampler, error) {
	return &BashSampler{ command: item.Path }, nil
}

func (s *BashSampler) Sample() (map[string]int64, error) {
	data, err := exec.Command("/bin/bash", "-c", s.command).Output()
	if err != nil {
		return nil, err
	}

	val, err := strconv.ParseInt(strings.Trim(string(data), "\n"), 10, 64)
	if err != nil {
		return nil, err
	}

	return map[string]int64{ "": val }, nil
}

