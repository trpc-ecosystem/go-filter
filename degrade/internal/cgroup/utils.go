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
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// readUint64FromFile 从文件中读取 uint64 类型内容
func readUint64FromFile(file string) (ret uint64, err error) {
	dat, err := readFromFile(file)
	if err != nil {
		return ret, err
	}

	ret, err = strconv.ParseUint(dat, 10, 64)
	return
}

// readInt64FromFile 从文件中度一个 int64 类型的数
func readInt64FromFile(file string) (ret int64, err error) {
	dat, err := readFromFile(file)
	if err != nil {
		return ret, err
	}

	ret, err = strconv.ParseInt(dat, 10, 64)
	return
}

// readMapFromFile 从文件中读取一系列数据
func readMapFromFile(name string) (ret map[string]uint64, err error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ret = make(map[string]uint64)
	scanner := bufio.NewScanner(file)
	// 逐行读数据
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)
		if len(items) != 2 {
			continue
		}
		v, err := strconv.ParseUint(strings.TrimSpace(items[1]), 10, 64)
		if err != nil {
			continue
		}
		ret[items[0]] = v
	}

	return ret, nil
}

// readFromFile 从文件中读取内容
func readFromFile(name string) (ret string, err error) {
	file, err := os.Open(name)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	return strings.TrimSpace(string(data)), nil
}
