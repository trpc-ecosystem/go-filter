// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package retry

import (
	"context"
	"time"

	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/naming/bannednodes"

	"trpc.group/trpc-go/trpc-filter/slime/throttle"
)

// ThrottledRetry defines a retry policy with throttle.
//
// A Retry should not be bound to a throttle. Instead, user may bind one throttle to some Retrys.
// This is why we introduce a new struct instead of adding a new field to Retry.
type ThrottledRetry struct {
	*Retry
	throttle throttle.Throttler
}

// NewThrottledRetry create a new ThrottledRetry.
func NewThrottledRetry(
	maxAttempts int,
	ecs []int,
	throttle throttle.Throttler,
	opts ...Opt,
) (*ThrottledRetry, error) {
	r, err := New(maxAttempts, ecs, opts...)
	if err != nil {
		return nil, err
	}

	return r.NewThrottledRetry(throttle), nil
}

// Invoke invokes handler f with Retry policy.
func (r *ThrottledRetry) Invoke(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) error {
	ctx = client.WithOptionsImmutable(ctx)
	if r.skipVisitedNodes == nil {
		ctx = bannednodes.NewCtx(ctx, false)
	} else if *r.skipVisitedNodes {
		ctx = bannednodes.NewCtx(ctx, true)
	}

	l := r.newLazyLog()

	impl := r.newImpl(ctx, req, rsp, f, l)
	impl.Start()

	if r.logCondition(impl) {
		l.Printf(impl.String())
		l.FlushCtx(ctx)
	}
	r.reporter.Report(ctx, impl)

	return impl.err
}

// newImpl create an impl from ThrottledRetry.
func (r *ThrottledRetry) newImpl(
	ctx context.Context,
	req, rsp interface{},
	handler filter.ClientHandleFunc,
	log logger,
) *impl {
	return &impl{
		ThrottledRetry: r,
		ctx:            ctx,
		req:            req,
		rsp:            rsp,
		handler:        handler,
		timer:          time.NewTimer(0), // start first attempt at once
		log:            log,
	}
}
