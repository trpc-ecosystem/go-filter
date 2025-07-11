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

package hedging

import (
	"context"
	"time"

	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/naming/bannednodes"

	"trpc.group/trpc-go/trpc-filter/slime/throttle"
)

// ThrottledHedging defines a hedging policy with throttler.
//
// A Hedging should not be bound to a throttle. Instead, user may bind one throttle to many Hedgings.
// This is why we introduce a new struct instead of adding a new field to Hedging.
type ThrottledHedging struct {
	*Hedging
	throttle throttle.Throttler
}

// NewThrottledHedging create a new ThrottledHedging.
func NewThrottledHedging(
	maxAttempts int,
	nonFatalECs []int,
	throttle throttle.Throttler,
	opts ...Opt,
) (*ThrottledHedging, error) {
	h, err := New(maxAttempts, nonFatalECs, opts...)
	if err != nil {
		return nil, err
	}

	return h.NewThrottledHedging(throttle), nil
}

// Invoke invokes handler f with hedging policy.
func (h *ThrottledHedging) Invoke(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) error {
	ctx = client.WithOptionsImmutable(ctx)
	if h.skipVisitedNodes == nil {
		ctx = bannednodes.NewCtx(ctx, false)
	} else if *h.skipVisitedNodes {
		ctx = bannednodes.NewCtx(ctx, true)
	}

	l := h.newLazyLog()

	impl := h.newImpl(ctx, req, rsp, f, l)
	impl.Start()

	if h.logCondition(impl) {
		l.Printf(impl.String())
		l.FlushCtx(ctx)
	}
	impl.reporter.Report(ctx, impl)

	return impl.err
}

// newImpl create an impl from ThrottledHedging.
func (h *ThrottledHedging) newImpl(
	ctx context.Context,
	req, rsp interface{},
	handler filter.ClientHandleFunc,
	log logger,
) *impl {
	return &impl{
		ThrottledHedging: h,
		ctx:              ctx,
		req:              req,
		rsp:              rsp,
		handler:          handler,
		results:          make(chan *attempt),
		timer:            time.NewTimer(0), // start first attempt at once.
		done:             make(chan struct{}),
		log:              log,
	}
}
