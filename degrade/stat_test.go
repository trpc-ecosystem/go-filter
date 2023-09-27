// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package degrade

import (
	"errors"
	"testing"
	"time"

	"github.com/shirou/gopsutil/load"
	"github.com/stretchr/testify/assert"
)

var errFake = errors.New("fake error")

// TestUpdateSysInfoPerTime test get sysinfo
func TestUpdateSysInfoPerTime(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"modified time"},
	}
	for _, tt := range tests {
		UpdateSysPeriod = 1
		WaitCPUTime = 1
		t.Run(tt.name, func(t *testing.T) {
			go UpdateSysInfoPerTime()
			time.Sleep(time.Second * time.Duration(2))
		})
	}
}

// TestGetMemoryStat test getmemorystat
func TestGetMemoryStat(t *testing.T) {
	memoryUsageInfosProvider = func() (usage float64, total uint64, rss uint64, err error) {
		return 100.000, 0, 0, nil
	}
	assert.Equal(t, float64(10000), GetMemoryStat())

	memoryUsageInfosProvider = func() (usage float64, total uint64, rss uint64, err error) {
		return 0, 0, 0, errFake
	}
	assert.Equal(t, 0.0, GetMemoryStat())
}

// TestGetCPUIdle cpu idle test
func TestGetCPUIdle(t *testing.T) {
	idle := GetCPUIdle()
	assert.Greater(t, idle, 0)
}

// TestGetLoadAvg load 测试
func TestGetLoadAvg(t *testing.T) {
	loadavgProvider = func() (*load.AvgStat, error) {
		return &load.AvgStat{}, nil
	}
	avg, err := GetLoadAvg()
	assert.NotNil(t, avg)
	assert.Nil(t, err)

	loadavgProvider = func() (*load.AvgStat, error) {
		return nil, errFake
	}
	avg, err = GetLoadAvg()
	assert.Equal(t, avg, &load.AvgStat{})
	assert.Equal(t, err, errFake)
}
