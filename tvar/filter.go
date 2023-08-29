// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package tvar 对rpc请求量、成功量、失败量、耗时分布、qps、百分位数进行统计上报
package tvar

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
)

const filterName = "tvar"

func init() {
	filter.Register(filterName, RPCServerFilter, RPCClientFilter)
}

// RPCServerFilter serverside rpc统计上报
func RPCServerFilter(
	ctx context.Context, req interface{}, handler filter.ServerHandleFunc,
) (rsp interface{}, err error) {
	cmd := trpc.Message(ctx).ServerRPCName()
	attr := attribute.String("method", cmd)

	serviceReqNum.Add(ctx, 1, attr)
	serviceReqActiveNum.Add(ctx, 1, attr)
	serviceQWin.Record()

	// TODO req包尺寸依赖：https://git.woa.com/trpc-go/trpc-go/issues/716
	//      rsp包尺寸filter里不方便统计
	// serviceReqSize.Record()

	begin := time.Now()
	defer func() {
		serviceReqActiveNum.Add(ctx, -1, attr)
		serviceRspNum.Add(ctx, 1, attr)

		if err != nil {
			serviceErrNum.Add(ctx, 1, attr)
			if isBusinessError(err) {
				serviceBusiErrNum.Add(ctx, 1, attr)
			}
		}

		d := time.Since(begin)
		serviceLatency.Record(ctx, d.Milliseconds(), attr)
	}()

	return handler(ctx, req)
}

// RPCClientFilter clientside rpc统计上报
func RPCClientFilter(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) (err error) {
	cmd := trpc.Message(ctx).ClientRPCName()
	attr := attribute.String("method", cmd)

	clientReqNum.Add(ctx, 1, attr)
	clientReqActiveNum.Add(ctx, 1, attr)

	begin := time.Now()
	defer func() {
		clientReqActiveNum.Add(ctx, -1, attr)
		clientRspNum.Add(ctx, 1, attr)
		if err != nil {
			clientErrNum.Add(ctx, 1, attr)
		}

		d := time.Since(begin)
		clientLatency.Record(ctx, d.Milliseconds(), attr)
	}()

	return handler(ctx, req, rsp)
}

func isBusinessError(err error) bool {
	e, ok := err.(*errs.Error)
	if !ok {
		return true
	}
	return e.Type == errs.ErrorTypeBusiness
}
