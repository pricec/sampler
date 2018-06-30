package samplers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"github.com/pricec/sampler/config"
)

type UptimeSampler struct {}

func NewUptimeSampler(item *config.ConfigItem) (*UptimeSampler, error) {
	return &UptimeSampler{}, nil
}

func (s *UptimeSampler) Sample() (map[string]int64, error) {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(data), " ")
	if len(parts) != 2 {
		return nil, errors.New(
			fmt.Sprintf(
				"Unexpected contents of /proc/uptime: %v",
				string(data),
			),
		)
	}

	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}

	return map[string]int64{ "": int64(val) }, nil
}
