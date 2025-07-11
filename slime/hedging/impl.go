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
	"errors"
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

// InflightErr should never be returned to client.
var InflightErr = errors.New("request is still inflight")

// impl contains some useful fields to implement hedging request.
type impl struct {
	*ThrottledHedging

	ctx     context.Context
	req     interface{}
	rsp     interface{}
	err     error
	handler filter.ClientHandleFunc

	inflightN int
	frozen    bool
	throttled bool

	cost     time.Duration
	results  chan *attempt
	attempts []*attempt

	timer *time.Timer
	done  chan struct{}

	log logger
}

// newAttempt create a new attempt.
//
// The msg and rsp in impl are copied to attempt.
// A child context with cancel is used to cancel inflight attempt when hedging finished.
// newAttempt freeze impl if all attempts has been drained or throttle check is failed.
func (impl *impl) newAttempt() (*attempt, error) {
	ctx, msg := codec.WithNewMessage(impl.ctx)
	if err := cpmsg.CopyMsg(msg, codec.Message(impl.ctx)); err != nil {
		return nil, fmt.Errorf("failed to create new attempt: %w", err)
	}
	ctx, cancel := context.WithCancel(ctx)

	a := attempt{
		impl:     impl,
		ctx:      ctx,
		cancel:   cancel,
		rsp:      copyutils.New(impl.rsp),
		inflight: true,
		attempt:  len(impl.attempts) + 1,
	}

	impl.attempts = append(impl.attempts, &a)
	impl.inflightN++

	impl.log.Printf("start %dth attempt, current inflightN: %d", a.attempt, impl.inflightN)

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

// Start start the main loop of hedging.
func (impl *impl) Start() {
	start := time.Now()
	defer func() {
		close(impl.done)
		for _, a := range impl.attempts {
			if a.inflight {
				a.Cancel()
			}
		}
		impl.cost = time.Since(start)
	}()

	for {
		select {
		case <-impl.ctx.Done():
			impl.log.Printf("hedging finished for timeout error")
			impl.err = TimeoutErr
			return
		case <-impl.timer.C:
			a, err := impl.newAttempt()
			if err != nil {
				impl.err = err
				return
			}
			a.AsyncStart()

			impl.scheduleNext(impl.hedgingDelay())
		case rst := <-impl.results:
			if impl.onReturn(rst) {
				impl.log.Printf("%dth attempt is return to client", rst.attempt)
				return
			}
		}
	}
}

// onReturn process the returned attempt.
//
// It returns a boolean indicate whether should the attempt terminate main loop of impl.
func (impl *impl) onReturn(a *attempt) (final bool) {
	impl.inflightN--
	impl.log.Printf("%dth attempt has returned, current inflightN: %d", a.attempt, impl.inflightN)

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

	if impl.isFatalErr(a.err) {
		return true
	}

	if a.pushbackDelay == nil {
		impl.scheduleNext(0)
	} else {
		impl.log.Printf("server issues a pushback delay: %v", *a.pushbackDelay)
		impl.scheduleNext(*a.pushbackDelay)
	}

	return impl.frozen && impl.inflightN == 0
}

// scheduleNext schedules next hedging request.
func (impl *impl) scheduleNext(delay time.Duration) {
	if impl.frozen {
		return
	}

	if delay < 0 {
		impl.timer.Stop()
		impl.frozen = true
		impl.log.Printf("freeze hedging for negative delay")
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
	for _, a := range impl.attempts {
		attempts = append(attempts, a)
	}
	return attempts
}

// Throttled implements view.Stat.
func (impl *impl) Throttled() bool {
	return impl.throttled
}

// InflightN implements view.Stat.
func (impl *impl) InflightN() int {
	return impl.inflightN
}

// Error implements view.Stat.
func (impl *impl) Error() error {
	return impl.err
}

// String implements fmt.Stringer.
func (impl *impl) String() string {
	var s string
	s += fmt.Sprintf("totalAttempts: %d, inflightN: %d, throttled: %t, finalErr: %v\n",
		len(impl.attempts), impl.inflightN, impl.throttled, impl.err)
	for _, a := range impl.attempts {
		s += "\t" + a.String() + "\n"
	}
	return s[:len(s)-1]
}

// attempt preserves the info the each attempt.
//
// err, end, pushbackDelay and msg in ctx are protected by inflight.
// They should only be accessed when inflight is false.
type attempt struct {
	*impl

	ctx    context.Context
	cancel func()
	rsp    interface{}
	err    error

	inflight      bool
	attempt       int
	start, end    time.Time
	pushbackDelay *time.Duration
}

// AsyncStart starts the attempt asynchronously.
func (a *attempt) AsyncStart() {
	a.start = time.Now()
	go func() {
		a.err = a.handler(a.ctx, a.req, a.rsp)
		a.end = time.Now()

		a.pushbackDelay = pushback.FromMsg(codec.Message(a.ctx))

		select {
		case <-a.impl.done:
			a.OnMissing()
		case a.results <- a:
		}
	}()
}

// OnReturn do some clean up and report.
func (a *attempt) OnReturn() {
	a.inflight = false
	a.cancel()
	a.ackThrottle()
}

// OnMissing means the attempt is too late to ack. This happens when another attempt has finished main loop of impl.
func (a *attempt) OnMissing() {
	codec.PutBackMessage(codec.Message(a.ctx))
}

// Cancel cancels the attempt.
//
// Cancel should be called immediately at the end of main loop of impl. However, OnMissing should only be called after
// the attempt has returned.
func (a *attempt) Cancel() {
	a.cancel()
	a.ackThrottle()
}

// ackThrottle ack the throttle with success or failure.
func (a *attempt) ackThrottle() {
	if !a.inflight && !a.noHedging() {
		if a.err == nil {
			a.throttle.OnSuccess()
			return
		}
		if a.isFatalErr() {
			return
		}
	}
	a.throttle.OnFailure()
}

func (a *attempt) isFatalErr() bool {
	return a.ThrottledHedging.isFatalErr(a.err)
}

func (a *attempt) noHedging() bool {
	return a.pushbackDelay != nil && *a.pushbackDelay < 0
}

// Start implements view.Attempt.
func (a *attempt) Start() time.Time {
	return a.start
}

// End implements view.Attempt.
func (a *attempt) End() time.Time {
	if a.inflight {
		return time.Time{}
	}
	return a.end
}

// Start implements view.Attempt.
func (a *attempt) Error() error {
	if a.inflight {
		return InflightErr
	}
	return a.err
}

// Inflight implements view.Attempt.
func (a *attempt) Inflight() bool {
	return a.inflight
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
	if a.inflight {
		return fmt.Sprintf("%dth attempt, start: %v, end: inflight, pushbackDelay: unknown, err: unknown",
			a.attempt, a.start.Format(timeFormat))
	}
	if a.pushbackDelay == nil {
		return fmt.Sprintf("%dth attempt, start: %v, end: %v, pushbackDelay: nil, err: %v",
			a.attempt, a.start.Format(timeFormat), a.end.Format(timeFormat), a.err)
	}
	if *a.pushbackDelay < 0 {
		return fmt.Sprintf("%dth attempt, start: %v, end: %v, pushbackDelay: no_hedging, err: %v",
			a.attempt, a.start.Format(timeFormat), a.end.Format(timeFormat), a.err)
	}
	return fmt.Sprintf("%dth attempt, start: %v, end: %v, pushbackDelay: %v, err: %v",
		a.attempt, a.start.Format(timeFormat), a.end.Format(timeFormat), a.pushbackDelay, a.err)
}
