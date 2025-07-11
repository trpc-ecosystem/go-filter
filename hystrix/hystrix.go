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

// Package hystrix is server fuse downgrade protection filter.
package hystrix

import (
	"context"
	"fmt"
	"runtime"

	"github.com/afex/hystrix-go/hystrix"
	metriccollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginType = "circuitbreaker"
	pluginName = "hystrix"
	filterName = "hystrix"
)

var (
	// WildcardKey is a wild card symbol.
	WildcardKey = "*"
	// ExcludeKey excludes prefix key.
	ExcludeKey = "_"

	cfg         map[string]hystrix.CommandConfig
	panicBufLen = 1024
)

type hystrixPlugin struct {
}

func init() {
	plugin.Register(pluginName, &hystrixPlugin{})
}

// Setup ...
func (p *hystrixPlugin) Setup(name string, decoder plugin.Decoder) error {
	cfg = make(map[string]hystrix.CommandConfig)
	err := decoder.Decode(cfg)
	if err != nil {
		log.Errorf("decoder.Decode(%T) err(%s)", cfg, err.Error())
		return err
	}
	hystrix.Configure(cfg)
	filter.Register(filterName, ServerFilter(), ClientFilter())
	return nil
}

// Type ...
func (p *hystrixPlugin) Type() string {
	return pluginType
}

// recoveryHandler prints call stack information.
var recoveryHandler = func(ctx context.Context, e interface{}) error {
	buf := make([]byte, panicBufLen)
	buf = buf[:runtime.Stack(buf, false)]
	log.ErrorContextf(ctx, "[Hystrix-Panic] %v\n%s\n", e, buf)
	return fmt.Errorf("%v", e)
}

// MetricCollectorFunc creates a MetricCollector function type definition.
type MetricCollectorFunc func(string) metriccollector.MetricCollector

// RegisterCollector is a register the statistical data collector.
// Call this function to register after implementing the collector according to its own actual business.
func RegisterCollector(collector MetricCollectorFunc) {
	metriccollector.Registry.Register(collector)
}

// ServerFilter fuses server request.
func ServerFilter() filter.ServerFilter {
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (interface{}, error) {
		// Get routing and configuration.
		cmd := trpc.Message(ctx).ServerRPCName()
		if _, ok := cfg[cmd]; !ok {
			// Configuration does not exist.
			// Whether there is wild card symbol.
			if _, ok := cfg[WildcardKey]; !ok {
				// 没有直接返回
				return handler(ctx, req)
			}
			// Whether to exclude parts when opening wildcards.
			if _, ok := cfg[ExcludeKey+cmd]; ok {
				return handler(ctx, req)
			}
			cmd = WildcardKey
		}

		var rsp interface{}
		return rsp, hystrix.Do(cmd, func() (err error) {
			defer func() {
				if errPanic := recover(); errPanic != nil {
					err = recoveryHandler(ctx, errPanic)
				}
			}()
			rsp, err = handler(ctx, req)
			return err
		}, nil)
	}
}

// ClientFilter fuses client request.
func ClientFilter() filter.ClientFilter {
	return func(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
		// Get routing and configuration.
		cmd := trpc.Message(ctx).ClientRPCName()
		if _, ok := cfg[cmd]; !ok {
			// Configuration does not exist.
			// Whether there is wild card symbol.
			if _, ok := cfg[WildcardKey]; !ok {
				// 没有直接返回
				return handler(ctx, req, rsp)
			}
			//  Whether to exclude parts when opening wildcards.
			if _, ok := cfg[ExcludeKey+cmd]; ok {
				return handler(ctx, req, rsp)
			}
			cmd = WildcardKey
		}
		return hystrix.Do(cmd, func() (err error) {
			defer func() {
				if errPanic := recover(); errPanic != nil {
					err = recoveryHandler(ctx, errPanic)
				}
			}()
			return handler(ctx, req, rsp)
		}, nil)
	}
}
