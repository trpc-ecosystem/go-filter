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

// Package retry defines the retry policy in a narrow sense.
package retry

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

// Retry retry policy.
type Retry struct {
	maxAttempts      int
	bf               backoff
	retryableECs     map[int]struct{}
	retryableErr     func(error) bool
	rspToErr         func(interface{}) error
	skipVisitedNodes *bool

	logCondition func(view.Stat) bool
	newLazyLog   func() lazyLogger
	reporter     reporter
}

// backoff is used to implement backoff priorities.
// customizedBackoff has the highest priority.
// exponentialBackoff comes second.
// linearBackoff has the lowest priority.
type backoff interface {
	backoff(attempt int) time.Duration
}

// Opt is the option function to modify Retry.
type Opt func(*Retry) error

// New create a Retry policy.
// An error will be returned if provided args cannot build a valid Retry.
func New(maxAttempts int, ecs []int, opts ...Opt) (*Retry, error) {
	if maxAttempts <= 0 {
		return nil, errors.New("maxAttempts must be positive")
	}

	if maxAttempts > MaximumAttempts {
		maxAttempts = MaximumAttempts
	}

	r := Retry{
		maxAttempts:  maxAttempts,
		retryableECs: make(map[int]struct{}),
		rspToErr:     func(interface{}) error { return nil },
		logCondition: func(view.Stat) bool { return false },
		newLazyLog:   func() lazyLogger { return &log.NoopLog{} },
		reporter:     &metrics.Noop{},
	}

	for _, ec := range ecs {
		r.retryableECs[ec] = struct{}{}
	}

	for _, opt := range opts {
		if err := opt(&r); err != nil {
			return nil, fmt.Errorf("failed to apply Retry Opt(s), err: %w", err)
		}
	}

	if r.bf == nil {
		return nil, errors.New("backoff is uninitialized")
	}

	if len(r.retryableECs) == 0 && r.retryableErr == nil {
		return nil, errors.New("one of retryableECs or retryableErr must be provided")
	}

	if r.retryableErr == nil {
		r.retryableErr = func(err error) bool { return false }
	}

	return &r, nil
}

// isRetryableErr checks whether the error is retryable.
func (r *Retry) isRetryableErr(err error) bool {
	if _, ok := r.retryableECs[int(errs.Code(err))]; ok {
		return true
	}

	return r.retryableErr(err)
}

// Invoke calls Invoke of ThrottledRetry with a Noop throttle.
func (r *Retry) Invoke(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) error {
	return r.NewThrottledRetry(throttle.NewNoop()).
		Invoke(ctx, req, rsp, f)
}

// NewThrottledRetry create a new ThrottledRetry from receiver Retry.
func (r *Retry) NewThrottledRetry(throttle throttle.Throttler) *ThrottledRetry {
	return &ThrottledRetry{Retry: r, throttle: throttle}
}

// WithRetryableErr allows user to register an additional function to check retryable errors.
func WithRetryableErr(retryableErr func(error) bool) Opt {
	return func(r *Retry) error {
		if retryableErr == nil {
			return errors.New("need a non-nil retryableErr")
		}

		r.retryableErr = retryableErr
		return nil
	}
}

// WithRspToErr allows user to register an additional function to convert rsp body errors.
func WithRspToErr(rspToErr func(interface{}) error) Opt {
	return func(r *Retry) error {
		if rspToErr == nil {
			return errors.New("need a non-nil rspToErr")
		}
		r.rspToErr = func(rsp interface{}) (err error) {
			defer func() {
				if rc := recover(); rc != nil {
					err = fmt.Errorf("retry rspToErr paniced: %v", rc)
				}
			}()
			return rspToErr(rsp)
		}
		return nil
	}
}

// WithExpBackoff set backoff strategy as exponential backoff.
func WithExpBackoff(
	initial, maximum time.Duration,
	multiplier int,
) Opt {
	return func(r *Retry) error {
		if _, ok := r.bf.(*customizedBackoff); ok {
			tlog.Trace("omit exponentialBackoff, since a customizedBackoff has already been set")
			return nil
		}

		bf, err := newExponentialBackoff(initial, maximum, multiplier)
		if err != nil {
			return fmt.Errorf("failed to create new exponentialBackoff, err: %w", err)
		}

		r.bf = bf
		return nil
	}
}

// WithLinearBackoff set backoff strategy as linear backoff.
func WithLinearBackoff(bfs ...time.Duration) Opt {
	return func(r *Retry) error {
		switch r.bf.(type) {
		case *customizedBackoff:
			tlog.Trace("omit linearBackoff, since a customizedBackoff has already been set")
			return nil
		case *exponentialBackoff:
			tlog.Trace("omit linearBackoff, since an exponentialBackoff has already been set")
			return nil
		default:
		}

		bf, err := newLinearBackoff(bfs...)
		if err != nil {
			return fmt.Errorf("failed to create new linearBackoff, err: %w", err)
		}

		r.bf = bf
		return nil
	}
}

// WithBackoff set a user defined backoff function.
func WithBackoff(backoff func(attempt int) time.Duration) Opt {
	return func(r *Retry) error {
		bf, err := newCustomizedBackoff(backoff)
		if err != nil {
			return fmt.Errorf("failed to create new customizedBackoff, err: %w", err)
		}

		r.bf = bf
		return nil
	}
}

// WithSkipVisitedNodes set whether to skip visited nodes in next retry request.
//
// The behavior depends on selector implementation.
// If skip is true, selector **must** always not return a visited node.
// If skip is false, selectors of each hedging request act absolutely independently.
// Without this Opt, as the default behavior, selector **should** try its best to return a non-visited node.
// If all nodes has been visited, it **may** returns a node as its wish.
func WithSkipVisitedNodes(skip bool) Opt {
	return func(r *Retry) error {
		r.skipVisitedNodes = &skip
		return nil
	}
}

// WithConditionalLog set a conditional log for retry policy.
// Only requests which meet the condition will be displayed.
func WithConditionalLog(l log.Logger, condition func(stat view.Stat) bool) Opt {
	return func(r *Retry) error {
		r.logCondition = condition
		r.newLazyLog = func() lazyLogger {
			return log.NewLazyLog(l)
		}
		return nil
	}
}

// WithConditionalCtxLog set a conditional log for retry policy.
// Only requests which meet the condition will be displayed.
func WithConditionalCtxLog(l log.CtxLogger, condition func(stat view.Stat) bool) Opt {
	return func(r *Retry) error {
		r.logCondition = condition
		r.newLazyLog = func() lazyLogger {
			return log.NewLazyCtxLog(l)
		}
		return nil
	}
}

// WithEmitter set the emitter for retry policy.
func WithEmitter(emitter metrics.Emitter) Opt {
	return func(r *Retry) error {
		r.reporter = metrics.NewReport(emitter)
		return nil
	}
}
