//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package degrade

import (
	"time"

	"github.com/shirou/gopsutil/load"
	"trpc.group/trpc-go/trpc-filter/degrade/internal/cgroup"
)

var (
	// UpdateSysPeriod 更新系统数据同步状态的时间周期
	UpdateSysPeriod = 30
	// WaitCPUTime cpu 使用率的平均值区间周期，单位 Second
	WaitCPUTime = 90
	cpuIdle     = 100
)

// UpdateSysInfoPerTime 更新系统数据给全局变量
func UpdateSysInfoPerTime() {
	for range time.Tick(time.Duration(UpdateSysPeriod) * time.Second) {
		// get cpuinfo perorid 90s
		cpuUsage, cerr := cgroup.GetDockerCPUUsage(time.Second * time.Duration(WaitCPUTime))
		if cerr == nil {
			// use more 1 to calculute cpu idle to integer
			cpuIdle = 100 - int(cpuUsage*100)
		}
		if cpuIdle < 0 {
			cpuIdle = 0
		}
	}
}

// GetCPUIdle 获取 cpu 使用率
func GetCPUIdle() int {
	return cpuIdle
}

var loadavgProvider = load.Avg

// GetLoadAvg 获取 load 负载，包含 load1，load5，load15
func GetLoadAvg() (*load.AvgStat, error) {
	loadavg, err := loadavgProvider()
	if err != nil {
		return &load.AvgStat{}, err
	}
	return loadavg, nil
}

var memoryUsageInfosProvider = cgroup.GetDockerMemoryUsageInfos

// GetMemoryStat 获取内存状态
func GetMemoryStat() float64 {
	usage, _, _, err := memoryUsageInfosProvider()
	if err != nil {
		return 0.0
	}
	return usage * 100
}
