// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package tvar 对rpc请求量、成功量、失败量、耗时分布、qps、百分位数进行统计上报
package tvar

import (
	"context"
	"errors"
	"testing"

	"git.code.oa.com/trpc-go/trpc-go"
	"git.code.oa.com/trpc-go/trpc-go/errs"
	"github.com/stretchr/testify/assert"
)

func TestRPCServerFilter(t *testing.T) {
	initMeterProvider()
	initRPCMetrics()

	ctx := trpc.BackgroundContext()

	gotRsp, err := RPCServerFilter(ctx, &struct{}{}, func(context.Context, interface{}) (interface{}, error) {
		return &struct{}{}, nil
	})
	assert.NotNil(t, gotRsp)
	assert.Nil(t, err)

	gotRsp, err = RPCServerFilter(ctx, &struct{}{}, func(context.Context, interface{}) (interface{}, error) {
		return nil, errors.New("fake error")
	})
	assert.Nil(t, gotRsp)
	assert.NotNil(t, err)

	gotRsp, err = RPCServerFilter(ctx, &struct{}{}, func(context.Context, interface{}) (interface{}, error) {
		return nil, errs.New(0, "fake error")
	})
	assert.Nil(t, gotRsp)
	assert.NotNil(t, err)
}

func TestRPCClientFilter(t *testing.T) {
	initMeterProvider()
	initRPCMetrics()

	ctx := trpc.BackgroundContext()

	err := RPCClientFilter(ctx, &struct{}{}, &struct{}{}, func(ctx context.Context, req, rsp interface{}) error {
		return nil
	})
	assert.Nil(t, err)

	err = RPCClientFilter(ctx, &struct{}{}, &struct{}{}, func(ctx context.Context, req, rsp interface{}) error {
		return errors.New("fake error")
	})
	assert.NotNil(t, err)
}
