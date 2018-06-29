package samplers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"github.com/pricec/sampler/config"
)

var memInfoFields = []string{
	"MemTotal",
	"MemFree",
	"MemAvailable",
	"Buffers",
	"Cached",
	"SwapCached",
	"Active",
	"Inactive",
	"Active(anon)",
	"Inactive(anon)",
	"Active(file)",
	"Inactive(file)",
	"Unevictable",
	"Mlocked",
	"SwapTotal",
	"SwapFree",
	"Dirty",
	"Writeback",
	"AnonPages",
	"Mapped",
	"Shmem",
	"Slab",
	"SReclaimable",
	"SUnreclaim",
	"KernelStack",
	"PageTables",
	"NFS_Unstable",
	"Bounce",
	"WritebackTmp",
	"CommitLimit",
	"Committed_AS",
	"VmallocTotal",
	"VmallocUsed",
	"VmallocChunk",
	"HardwareCorrupted",
	"AnonHugePages",
	"ShmemHugePages",
	"ShmemPmdMapped",
	"HugePages_Total",
	"HugePages_Free",
	"HugePages_Rsvd",
	"HugePages_Surp",
	"Hugepagesize",
	"DirectMap4k",
	"DirectMap2M",
}

var memNameMap = map[string]string{
	"MemAvailable": "available",
	"MemFree":      "free",
	"MemTotal":     "total",
	"Buffers":      "buffers",
	"Cached":       "cached",
}

type MemorySampler struct {}

func NewMemorySampler(item *config.ConfigItem) (*MemorySampler, error) {
	return &MemorySampler{}, nil
}

func (s *MemorySampler) Sample() (map[string]int64, error) {
	memInfo, err := getMemInfo()
	if err != nil {
		return nil, err
	}

	result := map[string]int64{}
	for key, statName := range memNameMap {
		result[statName] = memInfo[key]
	}
	return result, nil
}

func getMemInfo() (map[string]int64, error) {
	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	result := map[string]int64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 1 {
			val, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return nil, err
			}

			result[strings.Trim(fields[0], ":")] = val
		}
	}
	return result, checkMemInfo(result)
}

func checkMemInfo(info map[string]int64) error {
	for _, item := range memInfoFields {
		if _, ok := info[item]; !ok {
			return errors.New(
				fmt.Sprintf("/proc/meminfo missing field '%v'", item),
			)
		}
	}
	return nil
}
