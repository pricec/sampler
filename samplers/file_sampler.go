package samplers

import (
	"io/ioutil"
	"strconv"
	"strings"
	"github.com/pricec/sampler/config"
)

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
