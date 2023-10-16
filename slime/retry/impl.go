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

package retry

import (
	"context"
	"fmt"
	"time"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-utils/copyutils"

	"trpc.group/trpc-go/trpc-filter/slime/cpmsg"
	"trpc.group/trpc-go/trpc-filter/slime/pushback"
	"trpc.group/trpc-go/trpc-filter/slime/view"
)

const (
	timeFormat = "15:04:05.000"
)

// TimeoutErr request deadline exceeded.
var TimeoutErr = errs.NewFrameError(errs.RetClientTimeout, "request timeout")

// impl contains some useful fields to implement retry request.
type impl struct {
	*ThrottledRetry

	ctx     context.Context
	req     interface{}
	rsp     interface{}
	err     error
	handler filter.ClientHandleFunc

	cost      time.Duration
	throttled bool
	frozen    bool
	attempts  []*attempt

	timer *time.Timer

	log logger
}

// newAttempt creates a new attempt.
//
// The msg and rsp in impl are copied to attempt.
// newAttempt freeze impl if all attempts has been drained or throttle check is failed.
func (impl *impl) newAttempt() (*attempt, error) {
	ctx, msg := codec.WithNewMessage(impl.ctx)
	if err := cpmsg.CopyMsg(msg, codec.Message(impl.ctx)); err != nil {
		return nil, fmt.Errorf("failed to create new attempt: %w", err)
	}

	a := attempt{
		impl:    impl,
		ctx:     ctx,
		rsp:     copyutils.New(impl.rsp),
		attempt: len(impl.attempts) + 1,
	}

	impl.attempts = append(impl.attempts, &a)

	impl.log.Printf("start %dth attempt", a.attempt)

	if len(impl.attempts) == impl.maxAttempts || !impl.throttle.Allow() {
		if len(impl.attempts) == impl.maxAttempts {
			impl.log.Printf("freeze hedging for no more attempts")
		} else {
			impl.throttled = true
			impl.log.Printf("freeze hedging for throttle")
		}

		impl.frozen = true
	}

	return &a, nil
}

// Start start the main loop of retry.
func (impl *impl) Start() {
	start := time.Now()
	defer func() {
		impl.cost = time.Since(start)
	}()

	for {
		select {
		case <-impl.ctx.Done():
			impl.log.Printf("retry finished for timeout error")
			impl.err = TimeoutErr
			return
		case <-impl.timer.C:
			a, err := impl.newAttempt()
			if err != nil {
				impl.err = err
				return
			}
			a.SyncStart()
			if impl.onReturn(a) {
				impl.log.Printf("%dth attempt is return to client", a.attempt)
				return
			}
		}
	}
}

// onReturn process the returned attempt.
//
// It returns a boolean indicate whether should the attempt terminate main loop of impl.
func (impl *impl) onReturn(a *attempt) (final bool) {
	impl.log.Printf("%dth attempt has returned", a.attempt)
	a.OnReturn()

	defer func() {
		if final {
			if err := cpmsg.CopyMsg(codec.Message(impl.ctx), codec.Message(a.ctx)); err != nil {
				impl.err = fmt.Errorf("failed to copy back msg: %w, attempt err: %s", err, a.err)
			} else {
				impl.err = a.err
			}
		}
		codec.PutBackMessage(codec.Message(a.ctx))
	}()
	if a.err == nil {
		a.err = impl.rspToErr(a.rsp)
	}
	if a.err == nil {
		a.err = copyutils.ShallowCopy(impl.rsp, a.rsp)
		return true
	}
	if !impl.isRetryableErr(a.err) {
		return true
	}

	if a.pushbackDelay == nil {
		impl.scheduleNext(impl.bf.backoff(a.attempt))
	} else {
		impl.log.Printf("server issues a pushback delay: %v", *a.pushbackDelay)
		impl.scheduleNext(*a.pushbackDelay)
	}

	return impl.frozen
}

// scheduleNext schedules next retry request.
func (impl *impl) scheduleNext(delay time.Duration) {
	if impl.frozen {
		return
	}

	if delay < 0 {
		impl.timer.Stop()
		impl.frozen = true
		impl.log.Printf("freeze retry for negative delay")
		return
	}

	if !impl.timer.Stop() {
		select {
		case <-impl.timer.C:
		default:
		}
	}
	impl.timer.Reset(delay)
}

// Cost implements view.Stat.
func (impl *impl) Cost() time.Duration {
	return impl.cost
}

// Attempts implements view.Stat.
func (impl *impl) Attempts() []view.Attempt {
	attempts := make([]view.Attempt, 0, len(impl.attempts))
	for _, att := range impl.attempts {
		attempts = append(attempts, att)
	}
	return attempts
}

// Throttled implements view.Stat.
func (impl *impl) Throttled() bool {
	return impl.throttled
}

// InflightN implements view.Stat.
func (impl *impl) InflightN() int {
	return 0
}

// Error implements view.Stat.
func (impl *impl) Error() error {
	return impl.err
}

// String implements fmt.Stringer.
func (impl *impl) String() string {
	var s string
	s += fmt.Sprintf("totalAttempts: %d, throttled: %t, finalErr: %v\n",
		len(impl.attempts), impl.throttled, impl.err)
	for _, a := range impl.attempts {
		s += "\t" + a.String() + "\n"
	}
	return s[:len(s)-1]
}

// attempt preserves the info the each attempt.
type attempt struct {
	*impl

	ctx context.Context
	rsp interface{}
	err error

	attempt       int
	start, end    time.Time
	pushbackDelay *time.Duration
}

// SyncStart starts the attempt synchronously.
func (a *attempt) SyncStart() {
	a.start = time.Now()
	a.err = a.handler(a.ctx, a.req, a.rsp)
	a.end = time.Now()

	a.pushbackDelay = pushback.FromMsg(codec.Message(a.ctx))
}

// OnReturn just report to throttle.
func (a *attempt) OnReturn() {
	a.ackThrottle()
}

// ackThrottle ack the throttle with success or failure.
func (a *attempt) ackThrottle() {
	if !a.noRetry() {
		if a.err == nil {
			a.throttle.OnSuccess()
			return
		}
		if !a.isRetryableErr() {
			return
		}
	}
	a.throttle.OnFailure()
}

func (a *attempt) isRetryableErr() bool {
	return a.impl.isRetryableErr(a.err)
}

func (a *attempt) noRetry() bool {
	return a.pushbackDelay != nil && *a.pushbackDelay < 0
}

// Start implements view.Attempt.
func (a *attempt) Start() time.Time {
	return a.start
}

// End implements view.Attempt.
func (a *attempt) End() time.Time {
	return a.end
}

// Error implements view.Attempt.
func (a *attempt) Error() error {
	return a.err
}

// Inflight implements view.Attempt.
func (a *attempt) Inflight() bool {
	return false
}

// NoMoreAttempt implements view.NoMoreAttempt.
func (a *attempt) NoMoreAttempt() bool {
	if a.pushbackDelay == nil {
		return false
	}
	return *a.pushbackDelay < 0
}

// String implements fmt.Stringer.
func (a *attempt) String() string {
	if a.pushbackDelay == nil {
		return fmt.Sprintf("%dth attempt, start: %v, end: %v, pushbackDelay: nil, err: %v",
			a.attempt, a.start.Format(timeFormat), a.end.Format(timeFormat), a.err)
	}
	if *a.pushbackDelay < 0 {
		return fmt.Sprintf("%dth attempt, start: %v, end: %v, pushbackDelay: no_retry, err: %v",
			a.attempt, a.start.Format(timeFormat), a.end.Format(timeFormat), a.err)
	}
	return fmt.Sprintf("%dth attempt, start: %v, end: %v, pushbackDelay: %v, err: %v",
		a.attempt, a.start.Format(timeFormat), a.end.Format(timeFormat), a.pushbackDelay, a.err)
}
