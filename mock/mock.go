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

// Package mock implements interface mock calls through interceptors.
package mock

import (
	"context"
	"math/rand"
	"time"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"

	trpc "trpc.group/trpc-go/trpc-go"
)

type options struct {
	mocks []*Item
}

// Option set options.
type Option func(*options)

// WithMock set mock data.
func WithMock(mock *Item) Option {
	return func(opts *options) {
		opts.mocks = append(opts.mocks, mock)
	}
}

// ClientFilter set client request mock interceptor.
func ClientFilter(opts ...Option) filter.ClientFilter {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	rand.Seed(time.Now().Unix())

	return func(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
		msg := trpc.Message(ctx)

		for _, mock := range o.mocks {
			if mock.Method != "" && mock.Method != msg.ClientRPCName() {
				continue
			}

			if mock.Percent == 0 || rand.Intn(100) >= mock.Percent {
				// Triggered by percentage. For example, if 20%, the random number 0-99, only 0-19 will trigger.
				continue
			}

			if mock.Timeout {
				<-ctx.Done()
				return errs.NewFrameError(errs.RetClientTimeout, "mock filter: timeout")
			}
			if mock.Delay > 0 {
				select {
				case <-ctx.Done():
					return errs.NewFrameError(errs.RetClientTimeout, "mock filter: timeout during delay mock")
				case <-time.After(mock.delay):
				}
			}
			if mock.Retcode > 0 {
				return errs.New(mock.Retcode, mock.Retmsg)
			}
			if mock.Body != "" {
				if err := codec.Unmarshal(mock.Serialization, mock.data, rsp); err != nil {
					return errs.NewFrameError(errs.RetClientDecodeFail, "mock filter Unmarshal: "+err.Error())
				}
				return nil
			}
		}

		return handler(ctx, req, rsp)
	}
}
