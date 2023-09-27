// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package cgroup

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

// Doc: https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt

const (
	memFile     = "/sys/fs/cgroup/memory/memory.stat"
	procMemFile = "/proc/meminfo"
)

var (
	machineMemoryTotal, _ = getMachineMemoryTotal()
)

// GetDockerMemoryUsageInfos 获取容器的内存使用相关信息
func GetDockerMemoryUsageInfos() (usage float64, total, rss uint64, err error) {
	dockerMemInfo, err := readMapFromFile(memFile)
	if err != nil {
		return usage, total, rss, err
	}

	quotaMemory, ok := dockerMemInfo["hierarchical_memory_limit"]
	if !ok {
		return usage, total, rss, errors.New("Invalid hierarchical_memory_limit")
	}

	// 宿主机内存大于容器配额内存
	if machineMemoryTotal > quotaMemory {
		total = quotaMemory
	} else {
		total = machineMemoryTotal
	}

	rss = dockerMemInfo["total_rss"] + dockerMemInfo["total_mapped_file"]

	usage = float64(rss) / float64(total)
	return
}

// getMachineMemoryTotal 获取机器总内存
// "/proc/meminfo" 数据格式：
// MemTotal:       197298928 kB
// MemFree:         4111768 kB
// MemAvailable:   134355500 kB
// Buffers:          171064 kB
// Cached:         102828996 kB
func getMachineMemoryTotal() (total uint64, err error) {
	file, err := os.Open(procMemFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)
		if len(items) != 2 {
			continue
		}
		// 解析 MemTotal 部分
		if items[0] != "MemTotal:" {
			continue
		}
		items[1] = strings.TrimSpace(items[1])
		value := strings.TrimSuffix(items[1], "kB")
		value = strings.TrimSpace(value)
		total, err = strconv.ParseUint(value, 10, 64)
		total *= 1024
		if err != nil {
			return 0, err
		}
		break
	}

	return
}
