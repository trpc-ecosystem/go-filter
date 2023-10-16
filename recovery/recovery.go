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

// Package recovery is a tRPC filter used to recover the server side from a panic.
package recovery

import (
	"context"
	"fmt"
	"runtime"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"
)

func init() {
	filter.Register("recovery", ServerFilter(), nil)
}

// PanicBufLen is the size of the buffer for storing the panic call stack log. The default value as below.
var PanicBufLen = 4096

type options struct {
	rh RecoveryHandler
}

// Option sets Recovery option.
type Option func(*options)

// RecoveryHandler is the Recovery handle function.
// Deprecated: Use recovery.Handler instead.
type RecoveryHandler = Handler //nolint:revive

// Handler is the Recovery handle function.
type Handler func(ctx context.Context, err interface{}) error

// WithRecoveryHandler sets Recovery handle function.
func WithRecoveryHandler(rh Handler) Option {
	return func(opts *options) {
		opts.rh = rh
	}
}

var defaultRecoveryHandler = func(ctx context.Context, e interface{}) error {
	buf := make([]byte, PanicBufLen)
	buf = buf[:runtime.Stack(buf, false)]
	log.ErrorContextf(ctx, "[PANIC]%v\n%s\n", e, buf)
	metrics.IncrCounter("trpc.PanicNum", 1)
	return errs.NewFrameError(errs.RetServerSystemErr, fmt.Sprint(e))
}

var defaultOptions = &options{
	rh: defaultRecoveryHandler,
}

// ServerFilter adds the recovery filter to the server.
func ServerFilter(opts ...Option) filter.ServerFilter {
	o := defaultOptions
	for _, opt := range opts {
		opt(o)
	}
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (rsp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = o.rh(ctx, r)
			}
		}()

		return handler(ctx, req)
	}
}
