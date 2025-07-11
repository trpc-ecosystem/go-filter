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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
	trpc "trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const configInfo = `
plugins:
  circuitbreaker:
    degrade:
      load5: 5
      cpu_idle: 30
      memory_use_p : 60
      degrade_rate : 30
      max_concurrent_cnt : 1
      max_timeout_ms : 100
`

// FakeDecoder fake decoder
type FakeDecoder struct {
	err error
}

// Decode 解码
func (d *FakeDecoder) Decode(cfg interface{}) error {
	if d.err != nil {
		return d.err
	}
	return nil
}

// TestFilter_PluginType test PluginType
func TestFilter_PluginType(t *testing.T) {
	p := &Degrade{}
	assert.Equal(t, pluginType, p.Type())
}

// TestPlugin_Setup test plugin register setup
func TestPlugin_Setup(t *testing.T) {
	config := trpc.Config{}
	err := yaml.Unmarshal([]byte(configInfo), &config)
	assert.Nil(t, err)

	conf := config.Plugins[pluginType][pluginName]
	p := &Degrade{}
	err = p.Setup(pluginName, &plugin.YamlNodeDecoder{Node: &conf})
	assert.Nil(t, err)

	// decode 错误
	err = p.Setup(pluginName, &FakeDecoder{err: errors.New("fake error")})
	assert.NotNil(t, err)

	// degrade rate is 0
	cfg.DegradeRate = 0
	err = p.Setup(pluginName, &FakeDecoder{})
	assert.Nil(t, err)

	// degrade rate is 100
	cfg.DegradeRate = 100
	err = p.Setup(pluginName, &FakeDecoder{})
	assert.Nil(t, err)

	// interval 1s

	cfg.Interval = 1
	err = p.Setup(pluginName, &plugin.YamlNodeDecoder{Node: &conf})
	assert.Nil(t, err)
	time.Sleep(2 * time.Second)
}

// TestGetConfig 配置单测
func TestGetConfig(t *testing.T) {
	assert.NotNil(t, GetConfig())
}

// TestDegradeFilter 熔断插件单测
func TestDegradeFilter(t *testing.T) {
	testHandleFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &struct{}{}, nil
	}

	isDegrade = true
	cfg.DegradeRate = -1
	rsp, err := Filter(context.Background(), nil, testHandleFunc)
	assert.NotNil(t, err)
	assert.Nil(t, rsp)

	cfg.MaxConcurrentCnt = 0
	isDegrade = false
	rsp, err = Filter(context.Background(), nil, testHandleFunc)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)

	// 达到最大并发数，过载
	cfg.MaxConcurrentCnt = 1
	cfg.DegradeRate = 0
	isDegrade = false
	sema <- struct{}{}
	rsp, err = Filter(context.Background(), nil, testHandleFunc)
	assert.NotNil(t, err)
	assert.Nil(t, rsp)
}
