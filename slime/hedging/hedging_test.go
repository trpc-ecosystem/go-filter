// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package hedging_test

import (
	"context"
	"errors"
	"log"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	prom "trpc.group/trpc-go/trpc-filter/slime/view/metrics/prometheus"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/naming/bannednodes"

	"trpc.group/trpc-go/trpc-filter/slime/hedging"
	"trpc.group/trpc-go/trpc-filter/slime/pushback"
	"trpc.group/trpc-go/trpc-filter/slime/view"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/atomic"
)

const (
	hedgingDelay = time.Millisecond * 50
)

var EnableLog = false

type ConsoleLog struct{}

func (l *ConsoleLog) Println(s string) {
	log.Println(s)
}

type NoopLog struct{}

func (l *NoopLog) Println(context.Context, string) {}

func WithLog() hedging.Opt {
	if EnableLog {
		return hedging.WithConditionalLog(&ConsoleLog{}, func(view.Stat) bool { return true })
	}
	return func(*hedging.Hedging) {}
}

func TestNewMissingHedgingDelay(t *testing.T) {
	_, err := hedging.New(3, []int{int(errs.RetClientNetErr)})
	require.NotNil(t, err)
}

func TestNewMissingNonFatalECsAndNonFatalErr(t *testing.T) {
	_, err := hedging.New(3, []int{}, hedging.WithStaticHedgingDelay(time.Second))
	require.NotNil(t, err)

	_, err = hedging.New(3, []int{int(errs.RetClientNetErr)}, hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	_, err = hedging.New(3, []int{},
		hedging.WithStaticHedgingDelay(time.Second),
		hedging.WithNonFatalErr(func(err error) bool { return true }))
	require.Nil(t, err)

	_, err = hedging.New(3,
		[]int{int(errs.RetClientNetErr)},
		hedging.WithStaticHedgingDelay(time.Second),
		hedging.WithNonFatalErr(func(err error) bool { return true }))
	require.Nil(t, err)
}

func TestStartFirstAttemptImmediately(t *testing.T) {
	hedgingDelay := time.Millisecond * 500
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		WithLog())
	require.Nil(t, err)

	start := time.Now()
	err = h.Invoke(
		context.Background(), nil, nil,
		func(ctx context.Context, req, rsp interface{}) error {
			return nil
		})
	require.Nil(t, err)
	require.True(t, time.Now().Before(start.Add(hedgingDelay/2)))
}

func TestSlowResponse(t *testing.T) {
	// this test would fail in slow CPUs, make sure we have a large enough hedging delay.
	hedgingDelay := time.Millisecond * 200
	h, err := hedging.New(
		3, []int{},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		WithLog(),
		hedging.WithNonFatalErr(func(err error) bool { return true }))
	require.Nil(t, err)

	var attempts atomic.Int32
	toSleep := make(chan time.Duration, 3)
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		// the id for first, second or third attempt should be 1, 2 or 3.
		*rsp.(*int) = 4 - len(toSleep)
		d := <-toSleep
		time.Sleep(d)
		return nil
	}

	toSleep <- hedgingDelay * 4 // slept time for first attempt after start
	toSleep <- hedgingDelay * 2 // slept time for second attempt after start
	toSleep <- 0                // no sleep for third attempt after start
	var id int
	err = h.Invoke(context.Background(), nil, &id, handler)
	require.Nil(t, err)
	require.Equal(t, 3, id)
	require.Equal(t, int32(3), attempts.Load())

	// this time, the second attempt will be returned.
	attempts.Store(0)
	toSleep <- hedgingDelay * 4
	toSleep <- hedgingDelay * 2
	toSleep <- hedgingDelay * 3
	err = h.Invoke(context.Background(), nil, &id, handler)
	require.Nil(t, err)
	require.Equal(t, 2, id)
	require.Equal(t, int32(3), attempts.Load())
}

func TestImmediatelyHedgeOnRsp(t *testing.T) {
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(time.Second))
	require.Nil(t, err)

	var attempts atomic.Int32
	var starts [3]time.Time
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		starts[attempts.Load()-1] = time.Now()
		return errs.New(int(errs.RetClientNetErr), "")
	}

	err = h.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(3), attempts.Load())
	require.True(t, starts[1].Sub(starts[0]) < time.Millisecond*100)
	require.True(t, starts[2].Sub(starts[1]) < time.Millisecond*100)
}

func TestDeadlineExceeded(t *testing.T) {
	h, err := hedging.New(
		3, []int{},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		WithLog(),
		hedging.WithNonFatalErr(func(err error) bool { return true }))
	require.Nil(t, err)

	done := make(chan struct{})
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		<-done
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), hedgingDelay*10)
	err = h.Invoke(ctx, nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(3), attempts.Load())
	cancel()
	close(done)
}

func TestMaximumAttempts(t *testing.T) {
	h, err := hedging.New(
		3, []int{},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		WithLog(),
		hedging.WithNonFatalErr(func(err error) bool { return true }))
	require.Nil(t, err)

	done := make(chan struct{})
	attempts := atomic.Uint32{}
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		<-done
		return nil
	}

	rsp := make(chan error)
	go func() {
		rsp <- h.Invoke(context.Background(), nil, nil, handler)
	}()

	time.Sleep(time.Second)
	require.Equal(t, uint32(3), attempts.Load())

	close(done)
	require.Nil(t, <-rsp)
}

func TestNonFatalEC(t *testing.T) {
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	errsToReturn := make(chan error, 3)
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		return <-errsToReturn
	}

	errsToReturn <- errs.New(int(errs.RetClientNetErr), "")
	errsToReturn <- nil
	errsToReturn <- errs.New(errs.RetClientRouteErr, "")
	err = h.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn

	errsToReturn <- errs.New(int(errs.RetClientNetErr), "")
	errsToReturn <- errs.New(errs.RetClientRouteErr, "")
	errsToReturn <- nil
	attempts.Store(0)
	err = h.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.Equal(t, 1, len(errsToReturn))
}

func TestNonFatalErr(t *testing.T) {
	h, err := hedging.New(
		3, []int{},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		WithLog(),
		hedging.WithNonFatalErr(func(err error) bool {
			return strings.HasPrefix(err.Error(), "nonfatal")
		}))
	require.Nil(t, err)

	errsToReturn := make(chan error, 3)
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		return <-errsToReturn
	}

	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- nil
	errsToReturn <- errors.New("fatal")
	err = h.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn

	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- errors.New("fatal")
	errsToReturn <- nil
	attempts.Store(0)
	err = h.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn
}

func TestRspToErr(t *testing.T) {
	type Response struct {
		Ret int
	}
	h, err := hedging.New(
		4, []int{},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		WithLog(),
		hedging.WithNonFatalErr(func(err error) bool {
			return strings.HasPrefix(err.Error(), "nonfatal")
		}),
		hedging.WithRspToErr(func(rsp interface{}) error {
			switch v := rsp.(type) {
			case *Response:
				if v.Ret == 1 {
					return errors.New("nonfatal")
				} else if v.Ret == 2 {
					return errors.New("fatal")
				} else {
					return nil
				}
			}
			return nil
		}))
	require.Nil(t, err)

	errsToReturn := make(chan error, 4)
	retToReturn := make(chan int, 4)
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		v, ok := rsp.(*Response)
		if !ok {
			return nil
		}
		v.Ret = <-retToReturn
		return <-errsToReturn
	}

	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- nil
	retToReturn <- 1
	retToReturn <- 2
	rsp := &Response{}
	err = h.Invoke(context.Background(), nil, rsp, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.Equal(t, 0, len(errsToReturn))

	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- nil
	errsToReturn <- nil
	errsToReturn <- nil
	retToReturn <- 1
	retToReturn <- 1
	retToReturn <- 1
	retToReturn <- 3
	attempts.Store(0)
	err = h.Invoke(context.Background(), nil, rsp, handler)
	require.Nil(t, err)
	require.Equal(t, int32(4), attempts.Load())
	require.Equal(t, 0, len(errsToReturn))
	require.Equal(t, 3, rsp.Ret)
}

func TestServerPushback(t *testing.T) {
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	delay := hedgingDelay * 3
	nonFirst := atomic.Bool{}
	var attempts atomic.Int32
	starts := make([]time.Time, 2)
	handler := func(ctx context.Context, req, rsp interface{}) error {
		starts[attempts.Load()] = time.Now()
		attempts.Inc()

		if nonFirst.Load() {
			return nil
		}
		nonFirst.Store(true)

		msg := codec.Message(ctx)
		msg.ClientMetaData()[pushback.MetaKey] = []byte(delay.String())

		return errs.New(int(errs.RetClientNetErr), "")
	}

	err = h.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.True(t, starts[1].After(starts[0].Add(delay)))
}

func TestServerPushbackStop(t *testing.T) {
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		if attempts.Inc() == 2 {
			codec.Message(ctx).ClientMetaData()[pushback.MetaKey] = []byte(time.Duration(-1).String())
			return errs.New(int(errs.RetClientNetErr), "")
		}
		return errs.New(int(errs.RetClientNetErr), "")
	}

	err = h.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(2), attempts.Load())
}

func TestDynamicHedgingDelay(t *testing.T) {
	hedgingDelays := make(chan time.Duration, 1)
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithDynamicHedgingDelay(func() time.Duration {
			return <-hedgingDelays
		}))
	require.Nil(t, err)

	var attempts atomic.Int32
	starts := make([]time.Time, 3)
	handler := func(ctx context.Context, rsp, req interface{}) error {
		attempt := attempts.Inc()
		starts[attempt-1] = time.Now()
		hedgingDelays <- hedgingDelay * time.Duration(math.Pow(2, float64(attempt)))
		select {
		case <-ctx.Done():
			return errors.New("canceled")
		case <-time.After(time.Minute):
			panic("running too long")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	err = h.Invoke(ctx, nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(3), attempts.Load())
	require.True(t, hedgingDelay*2 < starts[1].Sub(starts[0]))
	require.True(t, hedgingDelay*4 < starts[2].Sub(starts[1]))
}

func TestCancelInFlightRequestOnReturn(t *testing.T) {
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	canceled := make(chan bool, 1)
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		if attempts.Inc() == 1 {
			// first attempts, sleep a long time and wait ctx to be canceled
			select {
			case <-ctx.Done():
				canceled <- true
				return errors.New("canceled")
			case <-time.After(hedgingDelay * 10):
				canceled <- false
				return nil
			}
		}

		return nil
	}

	err = h.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, int32(2), attempts.Load())
	require.True(t, <-canceled)
}

func TestClientCancels(t *testing.T) {
	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		<-ctx.Done()
		return errors.New("canceled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(hedgingDelay * 10)
		cancel()
	}()
	err = h.Invoke(ctx, nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(3), attempts.Load())
}

func TestConcurrencySafetyMsg(t *testing.T) {
	h, err := hedging.New(
		hedging.MaximumAttempts, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	begin := make(chan struct{})
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempt := strconv.Itoa(int(attempts.Inc()))
		<-begin
		codec.Message(ctx).ClientMetaData()[attempt] = []byte(attempt)
		*(rsp).(*string) = attempt
		return nil
	}

	go func() {
		time.Sleep(hedgingDelay * hedging.MaximumAttempts * 2)
		close(begin)
	}()

	var rsp string
	ctx := trpc.BackgroundContext()
	err = h.Invoke(ctx, nil, &rsp, handler)
	require.Nil(t, err)
	require.Equal(t, int32(hedging.MaximumAttempts), attempts.Load())
	msg := codec.Message(ctx)
	_, ok := msg.ClientMetaData()[rsp]
	require.True(t, ok)
	for i := 0; i < hedging.MaximumAttempts; i++ {
		attempt := strconv.Itoa(i)
		if attempt == rsp {
			continue
		}
		_, ok := msg.ClientMetaData()[attempt]
		require.False(t, ok)
	}
}

func TestConcurrencySafetyRsp(t *testing.T) {
	h, err := hedging.New(
		hedging.MaximumAttempts, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	type Rsp struct {
		data    map[string]string
		attempt string
	}

	begin := make(chan struct{})
	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempt := strconv.Itoa(int(attempts.Inc()))
		<-begin
		resp := rsp.(*Rsp)
		resp.data = make(map[string]string)
		resp.data[attempt] = attempt
		resp.attempt = attempt
		return nil
	}

	go func() {
		time.Sleep(hedgingDelay * hedging.MaximumAttempts * 2)
		close(begin)
	}()

	var rsp Rsp
	err = h.Invoke(context.Background(), nil, &rsp, handler)
	require.Nil(t, err)
	require.Equal(t, int32(hedging.MaximumAttempts), attempts.Load())
	require.Equal(t, 1, len(rsp.data))
	_, ok := rsp.data[rsp.attempt]
	require.True(t, ok)
}

func TestSkipVisitedNodes(t *testing.T) {
	type Expected struct {
		ok, mandatory bool
	}

	type Case struct {
		opt      []hedging.Opt
		expected Expected
	}

	cases := []Case{
		{
			[]hedging.Opt{WithLog()}, // default unset
			Expected{true, false},
		},
		{
			[]hedging.Opt{WithLog(), hedging.WithSkipVisitedNodes(false)},
			Expected{false, false},
		},
		{
			[]hedging.Opt{WithLog(), hedging.WithSkipVisitedNodes(true)},
			Expected{true, true},
		},
	}

	for _, cs := range cases {
		cs := cs

		opts := append(cs.opt, hedging.WithStaticHedgingDelay(hedgingDelay))
		h, err := hedging.New(
			hedging.MaximumAttempts, []int{int(errs.RetClientNetErr)},
			opts...)
		require.Nil(t, err)

		handler := func(ctx context.Context, req, rsp interface{}) error {
			_, mandatory, ok := bannednodes.FromCtx(ctx)
			assert.Equal(t, cs.expected.ok, ok)
			assert.Equal(t, cs.expected.mandatory, mandatory)
			return nil
		}

		err = h.Invoke(context.Background(), nil, nil, handler)
		require.Nil(t, err)
	}
}

func TestCopyBackMsg(t *testing.T) {
	fatalErr := errors.New("fatal error")
	nonFatalErr := errs.New(int(errs.RetClientNetErr), "")

	type Do func(context.Context) error

	okDo := func(ctx context.Context) error { return nil }
	delayedDo := func(ctx context.Context) error {
		time.Sleep(hedgingDelay * 3)
		return nil
	}
	newFailedDo := func(isFatal bool) func(ctx context.Context) error {
		var err error
		if isFatal {
			err = fatalErr
		} else {
			err = nonFatalErr
		}
		return func(ctx context.Context) error {
			return err
		}
	}
	noHedgingDo := func(ctx context.Context) error {
		codec.Message(ctx).ClientMetaData()[pushback.MetaKey] = []byte(time.Duration(-1).String())
		return nonFatalErr
	}

	type Expected struct {
		err   error
		reqID uint32
	}

	type Case struct {
		dos      []Do
		expected Expected
		msg      string
	}

	cases := []Case{
		{
			dos:      []Do{okDo},
			expected: Expected{nil, 1},
			msg:      "default case",
		},
		{
			dos:      []Do{delayedDo, okDo},
			expected: Expected{nil, 2},
			msg:      "first try return too slow",
		},
		{
			dos:      []Do{newFailedDo(false), newFailedDo(false), newFailedDo(false)},
			expected: Expected{nonFatalErr, 3},
			msg:      "all attempts failed",
		},
		{
			dos:      []Do{newFailedDo(false), newFailedDo(true), delayedDo},
			expected: Expected{fatalErr, 2},
			msg:      "second attempt encounter a fatal error",
		},
		{
			dos:      []Do{newFailedDo(false), noHedgingDo, delayedDo},
			expected: Expected{nonFatalErr, 2},
			msg:      "second attempt issue no hedging anymore",
		},
	}

	var attempts atomic.Uint32
	dos := make(chan func(context.Context) error, 3)
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempt := attempts.Add(1)
		codec.Message(ctx).WithRequestID(attempt)
		return (<-dos)(ctx)
	}

	h, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	for _, cs := range cases {
		attempts.Store(0)
	Drained:
		for {
			select {
			case <-dos:
			default:
				break Drained
			}
		}
		for _, do := range cs.dos {
			dos <- do
		}

		ctx := trpc.BackgroundContext()
		err := h.Invoke(ctx, nil, nil, handler)
		require.Equal(t, cs.expected.err, err, cs.msg)
		require.Equal(t, cs.expected.reqID, codec.Message(ctx).RequestID(), cs.msg)
	}
}

func TestLogAndEmitter(t *testing.T) {
	r, err := hedging.New(
		3, []int{int(errs.RetClientNetErr)},
		hedging.WithStaticHedgingDelay(hedgingDelay),
		hedging.WithConditionalCtxLog(&NoopLog{}, func(_ view.Stat) bool { return true }),
		hedging.WithEmitter(prom.NewEmitter()))
	require.Nil(t, err)

	require.Nil(t,
		r.Invoke(context.Background(), nil, nil,
			func(ctx context.Context, req, rsp interface{}) (err error) {
				return nil
			}))
}
