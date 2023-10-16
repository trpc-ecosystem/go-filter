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

// Package slime implement retry/hedging plugin.
//
// This package defines the way to config retry/hedging in yaml.
// You may use package retry/hedging directly if this package can not meet your requirements.
package slime

import (
	"context"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/filter"
)

// interceptor implement retry/hedging Filter.
func interceptor(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
	if disabled(ctx) {
		return handler(ctx, req, rsp)
	}

	rh := getRetryHedging(ctx)
	if rh == nil {
		return handler(ctx, req, rsp)
	}
	return rh.Invoke(ctx, req, rsp, handler)
}

// getRetryHedging get retryHedging from configuration.
// A nil return means retryHedging is not configured for this ctx.
func getRetryHedging(ctx context.Context) retryHedging {
	msg := codec.Message(ctx)

	calleeServiceName := msg.CalleeServiceName()
	if calleeServiceName == "" {
		return nil
	}

	service, ok := defaultManager.services[calleeServiceName]
	if !ok {
		return nil
	}

	calleeMethod := msg.CalleeMethod()
	if calleeMethod == "" {
		return service.retryHedging
	}

	if method, ok := service.methods[calleeMethod]; ok {
		return method
	}

	return service.retryHedging
}
