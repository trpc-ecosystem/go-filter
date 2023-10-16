//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package cgroup

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	cpuFile    = "/sys/fs/cgroup/cpuacct/cpuacct.usage_percpu"
	quotaFile  = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	periodFile = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	cpuSetFile = "/sys/fs/cgroup/cpuset/cpuset.cpus"

	cpuActUsageFile = "/sys/fs/cgroup/cpuacct/cpuacct.usage"
)

var (
	cpuTick, _ = getCPUTick()
	// cores 宿主机总核数
	cores, _ = GetCoreCount()
	// 容器分配可以使用的核数
	limitedCores, _ = GetLimitedCoreCount()

	errCores = errors.New("Error CPU Cores")
)

// GetDockerCPUUsage 获取 interval 时间间隔内容器 cpu 的利用率
func GetDockerCPUUsage(interval time.Duration) (usage float64, err error) {
	if interval <= 0 {
		return
	}

	preCPUTotal, err := GetContainerCPUTotal()
	if err != nil {
		return usage, err
	}

	// 这里阻塞 interval 时间
	time.Sleep(interval)
	postCPUTotal, err := GetContainerCPUTotal()
	if err != nil {
		return usage, err
	}

	usedCPU := float64(postCPUTotal - preCPUTotal)
	// 容器 cpu 配额比
	if cores == 0 {
		return usage, errCores
	}

	usage = usedCPU / (float64(interval) * limitedCores)

	return
}

// GetContainerCPUTotal 获取容器 cpu 使用时间
func GetContainerCPUTotal() (usage uint64, err error) {
	return readUint64FromFile(cpuActUsageFile)
}

// GetCoreCount 获取宿主机总共可用的 CPU 核数
func GetCoreCount() (num uint64, err error) {

	dat, err := readFromFile(cpuFile)
	if err != nil {
		return 0, err
	}

	items := strings.Split(dat, " ")
	return uint64(len(items)), nil
}

// GetLimitedCoreCount 获取容器 cpu 配额
func GetLimitedCoreCount() (mum float64, err error) {
	// 读取每个 period 时间间隔内可以使用的 cpu 时间
	quota, err := readInt64FromFile(quotaFile)
	if err != nil {
		return 0.0, err
	}

	// quota 为 -1 表示无限制，直接取可用的 cpu 核数
	if quota == -1 {
		return getValidCPUSet()
	}

	period, err := readInt64FromFile(periodFile)
	if err != nil {
		return 0.0, err
	}
	if period <= 0 {
		return 0.0, errors.New("invalid period num")
	}

	return float64(quota) / float64(period), nil
}

// getValidCPUSet 获取容器可以使用的 cpu 核
// 文件的内容存储格式例如：0,8-12,60-63
func getValidCPUSet() (ret float64, err error) {
	dat, err := readFromFile(cpuSetFile)
	if err != nil {
		return 0, err
	}

	validCores := 0
	items := strings.Split(dat, ",")
	for _, item := range items {
		cores := strings.Split(item, "-")

		switch len(cores) {
		case 1:
			// 直接指定某个核可用
			validCores++
		case 2:
			// 计算区间核数
			start, _ := strconv.Atoi(cores[0])
			end, _ := strconv.Atoi(cores[1])
			validCores += end - start + 1
		default:
			return 0.0, errors.New("Invalid cpu set formmat")
		}

	}

	return float64(validCores), nil
}

// getCPUTick 获取 CPU 调度周期（cpu 时钟）, 一般默认 100ms
func getCPUTick() (cnt int, err error) {
	out, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return cnt, err
	}

	cnt, err = strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return cnt, err
	}

	return
}
