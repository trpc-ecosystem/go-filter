// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package tvar

import (
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginType = "apm"
	pluginName = "tvar"
)

type config struct {
	Percentile []string `yaml:"percentile"`
}

var tvarCfg = config{}

func init() {
	plugin.Register(pluginName, &tvarFactory{})
}

type tvarFactory struct {
}

func (t tvarFactory) Type() string {
	return pluginType
}

var defaultLatencyPercentile = []string{"p50", "p99", "p999"}

func (t tvarFactory) Setup(name string, dec plugin.Decoder) error {
	if err := dec.Decode(&tvarCfg); err != nil {
		return err
	}
	if len(tvarCfg.Percentile) == 0 {
		tvarCfg.Percentile = defaultLatencyPercentile
	}
	initMeterProvider()
	initRPCMetrics()
	go start()
	return nil
}
