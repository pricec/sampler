package samplers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/scanner"
	"github.com/pricec/sampler/config"
)

var nameMap = []string{
	"user",
	"nice",
	"system",
	"idle",
	"iowait",
	"irq",
	"softirq",
	"steal",
	"guest",
	"guest_nice",
}

type CpuSampler struct {
	name string
}

func NewCpuSampler(item *config.ConfigItem) (*CpuSampler, error) {
	return &CpuSampler{
		name: item.Name,
	}, nil
}

func (s *CpuSampler) Sample() (map[string]int64, error) {
	data, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	result := map[string]int64{}
	// defer fmt.Printf("CpuSampler::Sample(): Result: %v\n", result)

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, s.name) {
			i := 0
			fields := strings.TrimPrefix(line, s.name)
			var s scanner.Scanner
			s.Init(strings.NewReader(fields))
			for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
				val, err := strconv.ParseInt(s.TokenText(), 10, 64)
				if err == nil {
					result[nameMap[i]] = val
				} else {
					return nil, err
				}
				i += 1
			}
			return result, nil
		}
	}

	return nil, errors.New(
		fmt.Sprintf("Failed to find stats for CPU '%v'", s.name),
	)
}
