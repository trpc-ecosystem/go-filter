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

// Package hedging defines hedging policy.
package hedging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	tlog "trpc.group/trpc-go/trpc-go/log"

	"trpc.group/trpc-go/trpc-filter/slime/throttle"
	"trpc.group/trpc-go/trpc-filter/slime/view"
	"trpc.group/trpc-go/trpc-filter/slime/view/log"
	"trpc.group/trpc-go/trpc-filter/slime/view/metrics"
)

const (
	// MaximumAttempts maximum attempts.
	MaximumAttempts = 5
)

type logger interface {
	Printf(string, ...interface{})
}

type lazyLogger interface {
	logger
	FlushCtx(context.Context)
}

type reporter interface {
	Report(context.Context, view.Stat)
}

// Hedging hedging policy.
type Hedging struct {
	maxAttempts      int
	hedgingDelay     func() time.Duration
	nonFatalECs      map[int]struct{}
	nonFatalErr      func(error) bool
	rspToErr         func(interface{}) error
	skipVisitedNodes *bool

	logCondition func(view.Stat) bool
	newLazyLog   func() lazyLogger
	reporter     reporter
}

// Opt is the option function to modify Hedging.
type Opt func(*Hedging)

// New create a new Hedging policy.
// An error will be returned if provided args cannot build a valid Hedging.
func New(maxAttempts int, nonFatalECs []int, opts ...Opt) (*Hedging, error) {
	if maxAttempts <= 0 {
		return nil, errors.New("maxAttempts must be positive")
	}

	if maxAttempts > MaximumAttempts {
		maxAttempts = MaximumAttempts
	}

	p := Hedging{
		maxAttempts:  maxAttempts,
		nonFatalECs:  make(map[int]struct{}),
		rspToErr:     func(interface{}) error { return nil },
		logCondition: func(view.Stat) bool { return false },
		newLazyLog:   func() lazyLogger { return &log.NoopLog{} },
		reporter:     &metrics.Noop{},
	}

	for _, ec := range nonFatalECs {
		p.nonFatalECs[ec] = struct{}{}
	}

	for _, opt := range opts {
		opt(&p)
	}

	if p.hedgingDelay == nil {
		return nil, errors.New("hedgingDelay is uninitialized")
	}

	if len(p.nonFatalECs) == 0 && p.nonFatalErr == nil {
		return nil, errors.New("one of nonFatalECs or nonFatalErr must be provided")
	}

	if p.nonFatalErr == nil {
		p.nonFatalErr = func(err error) bool { return false }
	}

	return &p, nil
}

// isFatalErr checks whether the error is fatal.
func (h *Hedging) isFatalErr(err error) bool {
	if _, ok := h.nonFatalECs[int(errs.Code(err))]; ok {
		return false
	}
	return !h.nonFatalErr(err)
}

// Invoke calls Invoke of ThrottledHedging with a Noop throttle.
func (h *Hedging) Invoke(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) error {
	return h.NewThrottledHedging(throttle.NewNoop()).
		Invoke(ctx, req, rsp, f)
}

// NewThrottledHedging create a new ThrottledHedging from receiver Hedging.
func (h *Hedging) NewThrottledHedging(throttle throttle.Throttler) *ThrottledHedging {
	return &ThrottledHedging{Hedging: h, throttle: throttle}
}

// WithStaticHedgingDelay set a static hedging delay for Hedging.
// WithStaticHedgingDelay should not be used along with WithDynamicHedgingDelay.
func WithStaticHedgingDelay(delay time.Duration) Opt {
	return func(h *Hedging) {
		h.hedgingDelay = func() time.Duration {
			return delay
		}
	}
}

// WithDynamicHedgingDelay set hedging delay with user defined function.
// WithDynamicHedgingDelay should not be used along with WithStaticHedgingDelay.
func WithDynamicHedgingDelay(f func() time.Duration) Opt {
	return func(h *Hedging) {
		h.hedgingDelay = f
	}
}

// WithNonFatalErr allows user to register an additional function to check fatal errors.
// A fatal error would cause the Call function to return immediately and ignores in-flight hedging requests and
// remaining attempts.
func WithNonFatalErr(nonFatalErr func(error) bool) Opt {
	return func(h *Hedging) {
		if nonFatalErr == nil {
			tlog.Trace("ignore nil nonFatalErr")
			return
		}

		h.nonFatalErr = nonFatalErr
	}
}

// WithRspToErr allows user to register an additional function to convert rsp body error
func WithRspToErr(rspToErr func(interface{}) error) Opt {
	return func(h *Hedging) {
		if rspToErr == nil {
			tlog.Trace("ignore nil rspToErr")
			return
		}
		h.rspToErr = func(rsp interface{}) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("hedging rspToErr paniced: %v", r)
				}
			}()
			return rspToErr(rsp)
		}
	}
}

// WithSkipVisitedNodes set whether to skip visited nodes in next hedging request.
//
// The behavior depends on selector implementation.
// If skip is true, selector **must** always not return a visited node.
// If skip is false, selectors of each hedging request act absolutely independently.
// Without this Opt, as the default behavior, selector **should** try its best to return a non-visited node.
// If all nodes has been visited, it **may** returns a node as its wish.
func WithSkipVisitedNodes(skip bool) Opt {
	return func(h *Hedging) {
		h.skipVisitedNodes = &skip
	}
}

// WithConditionalLog set a conditional log for hedging policy.
// Only requests which meet the condition will be displayed.
func WithConditionalLog(l log.Logger, condition func(stat view.Stat) bool) Opt {
	return func(h *Hedging) {
		h.logCondition = condition
		h.newLazyLog = func() lazyLogger {
			return log.NewLazyLog(l)
		}
	}
}

// WithConditionalCtxLog set a conditional log for hedging policy.
// Only requests which meet the condition will be displayed.
func WithConditionalCtxLog(l log.CtxLogger, condition func(stat view.Stat) bool) Opt {
	return func(h *Hedging) {
		h.logCondition = condition
		h.newLazyLog = func() lazyLogger {
			return log.NewLazyCtxLog(l)
		}
	}
}

// WithEmitter set the reporter for hedging policy.
func WithEmitter(emitter metrics.Emitter) Opt {
	return func(h *Hedging) {
		h.reporter = metrics.NewReport(emitter)
	}
}
