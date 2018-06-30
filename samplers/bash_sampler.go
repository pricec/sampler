package samplers

import (
	"os/exec"
	"strconv"
	"strings"
	"github.com/pricec/sampler/config"
)

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
