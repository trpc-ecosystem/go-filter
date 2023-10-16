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

package hystrix

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	metriccollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const configInfo = `
plugins:                                          # Plugin configuration.
  circuitbreaker:
    hystrix:
      /api/test:
        timeout: 8
        maxconcurrentrequests: 100
        requestvolumethreshold: 30
        sleepwindow: 5
`

func testOKServerHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func testOKHandler(ctx context.Context, req interface{}, rsp interface{}) error {
	return nil
}

func testPanicServerHandler(ctx context.Context, req interface{}) (interface{}, error) {
	panic("xxx")
}

func testPanicHandler(ctx context.Context, req interface{}, rsp interface{}) error {
	panic("xxx")
}

func testTimeoutServerHandler(ctx context.Context, req interface{}) (interface{}, error) {
	time.Sleep(20 * time.Millisecond)
	return nil, nil
}

func testTimeoutHandler(ctx context.Context, req interface{}, rsp interface{}) error {
	time.Sleep(20 * time.Millisecond)
	return nil
}

func testErrorServerHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errors.New("rpc error")
}

func testErrorHandler(ctx context.Context, req interface{}, rsp interface{}) error {
	return errors.New("rpc error")
}

func TestFilter_PluginType(t *testing.T) {
	p := &hystrixPlugin{}
	assert.Equal(t, pluginType, p.Type())
}

func TestPlugin_Setup(t *testing.T) {
	cfg := trpc.Config{}
	err := yaml.Unmarshal([]byte(configInfo), &cfg)
	assert.Nil(t, err)

	conf := cfg.Plugins[pluginType][pluginName]
	p := &hystrixPlugin{}
	err = p.Setup(pluginName, &plugin.YamlNodeDecoder{Node: &conf})
	assert.Nil(t, err)
}

func TestFilterFnuc(t *testing.T) {
	f := ServerFilter()
	ctx := trpc.BackgroundContext()
	trpc.Message(ctx).WithServerRPCName("/api/test")
	cfg = make(map[string]hystrix.CommandConfig)
	cfg["/api/test"] = hystrix.CommandConfig{}
	_, err := f(ctx, []byte("req"), testOKServerHandler)
	assert.Nil(t, err)

	_, err = f(ctx, []byte("req"), testPanicServerHandler)
	assert.NotNil(t, err)
}

func TestClientFilter(t *testing.T) {
	f := ClientFilter()
	ctx := trpc.BackgroundContext()
	cfg = make(map[string]hystrix.CommandConfig)
	cfg["/api/clientTest"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	hystrix.Configure(cfg)
	trpc.Message(ctx).WithClientRPCName("/api/clientTest")
	// Test Panic
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testPanicHandler)
		assert.NotNil(t, err)
	}
	// Request OK
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testOKHandler)
		assert.Nil(t, err)
	}
	// Request Timeout
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testTimeoutHandler)
		assert.EqualError(t, err, "hystrix: timeout")
	}
	// Circuit is opened for reqNum >= cfg.RequestVolumeThreshold && errNum > cfg.ErrorPercentThreshold
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testErrorHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
	// Request is rejected, because circuit is open
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testOKHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
	// Allowing one request but getting error again, circuit keeps open
	{
		time.Sleep(3 * time.Millisecond)
		err := f(ctx, []byte("req"), []byte("rsp"), testErrorHandler)
		assert.EqualError(t, err, "rpc error")
		err = f(ctx, []byte("req"), []byte("rsp"), testOKHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
}

func TestWildcardKeyClientFilter(t *testing.T) {
	f := ClientFilter()
	cfg = make(map[string]hystrix.CommandConfig)
	// Turn on global configuration.
	ctx := trpc.BackgroundContext()
	cfg["*"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	hystrix.Configure(cfg)
	trpc.Message(ctx).WithClientRPCName("/api/testNil")
	// Test Panic
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testPanicHandler)
		assert.NotNil(t, err)
	}
	// Request OK
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testOKHandler)
		assert.Nil(t, err)
	}
	// Request Timeout
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testTimeoutHandler)
		assert.EqualError(t, err, "hystrix: timeout")
	}
	// Circuit is opened for reqNum >= cfg.RequestVolumeThreshold && errNum > cfg.ErrorPercentThreshold
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testErrorHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
	// Request is rejected, because circuit is open
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testOKHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
}

func TestExcludeKeyClientFilter(t *testing.T) {
	f := ClientFilter()
	cfg = make(map[string]hystrix.CommandConfig)
	// Turn on global configuration.
	ctx := trpc.BackgroundContext()
	cfg["*"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	// Set to exclude this api.
	cfg["_/api/exclude"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	hystrix.Configure(cfg)

	trpc.Message(ctx).WithClientRPCName("/api/exclude")
	// Request OK
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testOKHandler)
		assert.Nil(t, err)
	}
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testTimeoutHandler)
		assert.Nil(t, err)
	}
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testErrorHandler)
		assert.EqualError(t, err, "rpc error")
	}
	{
		err := f(ctx, []byte("req"), []byte("rsp"), testErrorHandler)
		assert.EqualError(t, err, "rpc error")
	}
}

func TestWildcardKeyServerFilter(t *testing.T) {
	f := ServerFilter()
	cfg = make(map[string]hystrix.CommandConfig)
	// Turn on global configuration.
	ctx := trpc.BackgroundContext()
	WildcardKey = "**"
	cfg["**"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	hystrix.Configure(cfg)
	trpc.Message(ctx).WithServerRPCName("/api/testNil")
	// Test Panic
	{
		_, err := f(ctx, []byte("req"), testPanicServerHandler)
		assert.NotNil(t, err)
	}
	// Request OK
	{
		_, err := f(ctx, []byte("req"), testOKServerHandler)
		assert.Nil(t, err)
	}
	// Request Timeout
	{
		_, err := f(ctx, []byte("req"), testTimeoutServerHandler)
		assert.EqualError(t, err, "hystrix: timeout")
	}
	// Circuit is opened for reqNum >= cfg.RequestVolumeThreshold && errNum > cfg.ErrorPercentThreshold
	{
		_, err := f(ctx, []byte("req"), testErrorServerHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
	// Request is rejected, because circuit is open
	{
		_, err := f(ctx, []byte("req"), testOKServerHandler)
		assert.EqualError(t, err, "hystrix: circuit open")
	}
}

func TestExcludeKeyServerFilter(t *testing.T) {
	f := ServerFilter()
	cfg = make(map[string]hystrix.CommandConfig)
	// Turn on global configuration.
	ctx := trpc.BackgroundContext()
	cfg["*"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	// Set to exclude this api.
	cfg["_/api/exclude"] = hystrix.CommandConfig{
		Timeout:                10,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 3,
		SleepWindow:            3,
		ErrorPercentThreshold:  10,
	}
	hystrix.Configure(cfg)

	trpc.Message(ctx).WithServerRPCName("/api/exclude")
	// Request OK
	{
		_, err := f(ctx, []byte("req"), testOKServerHandler)
		assert.Nil(t, err)
	}
	{
		_, err := f(ctx, []byte("req"), testTimeoutServerHandler)
		assert.Nil(t, err)
	}
	{
		_, err := f(ctx, []byte("req"), testErrorServerHandler)
		assert.EqualError(t, err, "rpc error")
	}
	{
		_, err := f(ctx, []byte("req"), testErrorServerHandler)
		assert.EqualError(t, err, "rpc error")
	}
}

type testMetricCollector struct{}

// Update ...
func (m *testMetricCollector) Update(r metriccollector.MetricResult) {}

// Reset ...
func (m *testMetricCollector) Reset() {}

func newTestMetircCollector(name string) metriccollector.MetricCollector {
	return &testMetricCollector{}
}

func TestRegisterCollector(t *testing.T) {
	RegisterCollector(newTestMetircCollector)
}
