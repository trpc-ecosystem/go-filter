// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package retry_test

import (
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/naming/bannednodes"

	"trpc.group/trpc-go/trpc-filter/slime/pushback"
	"trpc.group/trpc-go/trpc-filter/slime/retry"
	"trpc.group/trpc-go/trpc-filter/slime/view"
	prom "trpc.group/trpc-go/trpc-filter/slime/view/metrics/prometheus"
)

const (
	retryBackoff = time.Millisecond * 50
)

var EnableLog = false

type ConsoleLog struct{}

func (l *ConsoleLog) Println(s string) {
	log.Println(s)
}

type NoopLog struct{}

func (l *NoopLog) Println(context.Context, string) {}

func WithLog() retry.Opt {
	if EnableLog {
		return retry.WithConditionalLog(&ConsoleLog{}, func(view.Stat) bool { return true })
	}
	return func(*retry.Retry) error { return nil }
}

func TestNewMissingBackoff(t *testing.T) {
	_, err := retry.New(3, []int{int(errs.RetClientNetErr)})
	require.NotNil(t, err)
}

func TestWithExpBackoff(t *testing.T) {
	_, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		retry.WithExpBackoff(-time.Second, time.Second, 2))
	require.NotNil(t, err, "initial backoff should be non negative")

	_, err = retry.New(
		3, []int{int(errs.RetClientNetErr)},
		retry.WithExpBackoff(time.Second, time.Millisecond, 2))
	require.NotNil(t, err, "initial backoff should be lesser than maximum")

	_, err = retry.New(
		3, []int{int(errs.RetClientNetErr)},
		retry.WithExpBackoff(time.Second, time.Second*2, 0))
	require.NotNil(t, err, "multiplier should be greater than zero")

	_, err = retry.New(
		3, []int{int(errs.RetClientNetErr)},
		retry.WithExpBackoff(time.Millisecond, time.Second, 2))
	require.Nil(t, err)
}

func TestStartFirstAttemptImmediately(t *testing.T) {
	retryBackoff := time.Millisecond * 500
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		return nil
	}

	start := time.Now()
	err = r.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, int32(1), attempts.Load())
	require.True(t, time.Now().Before(start.Add(retryBackoff/2)))
}

func TestDeadlineExceeded(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff * 10 }))
	require.Nil(t, err)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Inc()
		return errs.New(errs.RetClientNetErr, "")
	}

	ctx, cancel := context.WithTimeout(context.Background(), retryBackoff)
	err = r.Invoke(ctx, nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(1), attempts.Load())
	cancel()
}

func TestClientCancel(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	var attempts int
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts++
		if attempts == 2 {
			<-ctx.Done()
			return errors.New("canceled")
		}
		return errs.New(errs.RetClientNetErr, "")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(retryBackoff * 5)
		cancel()
	}()
	err = r.Invoke(ctx, nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, 2, attempts)
}

func TestMaxAttempts(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	var attempts int
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts++
		return errs.New(errs.RetClientNetErr, "")
	}

	err = r.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, 3, attempts)
}

func TestRetryableEC(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	errsToReturn := make(chan error, 3)
	var attempts int
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts++
		return <-errsToReturn
	}

	errsToReturn <- errs.New(errs.RetClientNetErr, "")
	errsToReturn <- nil
	errsToReturn <- errs.New(errs.RetClientRouteErr, "")
	err = r.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, 2, attempts)
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn

	errsToReturn <- errs.New(errs.RetClientNetErr, "")
	errsToReturn <- errs.New(errs.RetClientRouteErr, "")
	errsToReturn <- nil
	attempts = 0
	err = r.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, 2, attempts)
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn
}

func TestRetryableErr(t *testing.T) {
	r, err := retry.New(
		3, []int{},
		WithLog(),
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff }),
		retry.WithRetryableErr(func(err error) bool {
			return strings.HasPrefix(err.Error(), "nonfatal")
		}))
	require.Nil(t, err)

	errsToReturn := make(chan error, 3)
	var attempts int
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts++
		return <-errsToReturn
	}

	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- nil
	errsToReturn <- errors.New("fatal")
	err = r.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, 2, attempts)
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn

	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- errors.New("fatal")
	errsToReturn <- nil
	attempts = 0
	err = r.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, 2, attempts)
	require.Equal(t, 1, len(errsToReturn))
	<-errsToReturn
}

func TestRspToErr(t *testing.T) {
	type Response struct {
		Ret int
	}
	r, err := retry.New(
		4, []int{},
		WithLog(),
		retry.WithBackoff(func(attempt int) time.Duration { return retryBackoff }),
		retry.WithRetryableErr(func(err error) bool { return true }),
		retry.WithRspToErr(func(rsp interface{}) error {
			switch v := rsp.(type) {
			case *Response:
				if v.Ret == 1 || v.Ret == 2 {
					return errs.New(v.Ret, "failed")
				}
				return nil
			}
			return nil
		}))
	require.Nil(t, err)

	errsToReturn := make(chan error, 4)
	errsToReturn <- errors.New("nonfatal")
	errsToReturn <- nil
	errsToReturn <- nil
	errsToReturn <- nil

	retToReturn := make(chan int, 4)
	retToReturn <- 1
	retToReturn <- 2
	retToReturn <- 3
	retToReturn <- 4
	var attempts int
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts++
		v, ok := rsp.(*Response)
		if !ok {
			return nil
		}
		v.Ret = <-retToReturn
		return <-errsToReturn
	}
	rsp := &Response{}
	err = r.Invoke(context.Background(), nil, rsp, handler)
	require.Nil(t, err)
	require.Equal(t, 3, rsp.Ret)
	require.Equal(t, 3, attempts)
	require.Equal(t, 1, len(errsToReturn))
	require.Equal(t, 1, len(retToReturn))
}

func TestServerPushback(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	delay := retryBackoff * 10
	var attempts int
	starts := make([]time.Time, 2)
	handler := func(ctx context.Context, req, rsp interface{}) error {
		starts[attempts] = time.Now()
		attempts++

		if attempts > 1 {
			return nil
		}

		codec.Message(ctx).ClientMetaData()[pushback.MetaKey] = []byte(delay.String())

		return errs.New(errs.RetClientNetErr, "")
	}

	ctx := trpc.BackgroundContext()
	codec.Message(ctx).WithClientMetaData(make(codec.MetaData))
	err = r.Invoke(ctx, nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, 2, attempts)
	require.True(t, starts[1].After(starts[0].Add(delay)))
}

func TestServerPushbackStop(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	var attempts int
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts++
		if attempts == 2 {
			codec.Message(ctx).ClientMetaData()[pushback.MetaKey] = []byte(time.Duration(-1).String())
			return errs.New(errs.RetClientNetErr, "")
		}
		return errs.New(errs.RetClientNetErr, "")
	}

	ctx := trpc.BackgroundContext()
	codec.Message(ctx).WithClientMetaData(make(codec.MetaData))
	err = r.Invoke(ctx, nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, errs.RetClientNetErr, errs.Code(err))
	require.Equal(t, 2, attempts)
}

func TestSkipVisitedNodes(t *testing.T) {
	type Expected struct {
		ok, mandatory bool
	}

	type Case struct {
		opt      []retry.Opt
		expected Expected
	}

	cases := []Case{
		{
			[]retry.Opt{WithLog()}, // default unset
			Expected{true, false},
		},
		{
			[]retry.Opt{WithLog(), retry.WithSkipVisitedNodes(false)},
			Expected{false, false},
		},
		{
			[]retry.Opt{WithLog(), retry.WithSkipVisitedNodes(true)},
			Expected{true, true},
		},
	}

	for _, cs := range cases {
		cs := cs

		opts := append(cs.opt, retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
		r, err := retry.New(
			retry.MaximumAttempts, []int{int(errs.RetClientNetErr)},
			opts...)
		require.Nil(t, err)

		handler := func(ctx context.Context, req, rsp interface{}) error {
			_, mandatory, ok := bannednodes.FromCtx(ctx)
			assert.Equal(t, cs.expected.ok, ok)
			assert.Equal(t, cs.expected.mandatory, mandatory)
			return nil
		}

		err = r.Invoke(context.Background(), nil, nil, handler)
		require.Nil(t, err)
	}
}

func TestCopyBackMsg(t *testing.T) {
	nonRetryableErr := errors.New("fatal error")
	retryableErr := errs.New(errs.RetClientNetErr, "")

	type Do func(context.Context) error

	okDo := func(ctx context.Context) error { return nil }
	newFailedDo := func(retryable bool) func(ctx context.Context) error {
		var err error
		if retryable {
			err = retryableErr
		} else {
			err = nonRetryableErr
		}
		return func(ctx context.Context) error {
			return err
		}
	}
	noHedgingDo := func(ctx context.Context) error {
		codec.Message(ctx).ClientMetaData()[pushback.MetaKey] = []byte(time.Duration(-1).String())
		return retryableErr
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
			dos:      []Do{newFailedDo(true), newFailedDo(true), newFailedDo(true)},
			expected: Expected{retryableErr, 3},
			msg:      "all attempts failed",
		},
		{
			dos:      []Do{newFailedDo(true), newFailedDo(false)},
			expected: Expected{nonRetryableErr, 2},
			msg:      "second attempt encounter a non-retryable error",
		},
		{
			dos:      []Do{newFailedDo(true), noHedgingDo},
			expected: Expected{retryableErr, 2},
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

	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
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
		err := r.Invoke(ctx, nil, nil, handler)
		require.Equal(t, cs.expected.err, err, cs.msg)
		require.Equal(t, cs.expected.reqID, codec.Message(ctx).RequestID(), cs.msg)
	}
}

func TestLogAndEmitter(t *testing.T) {
	r, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		retry.WithLinearBackoff(time.Millisecond*10),
		retry.WithConditionalCtxLog(&NoopLog{}, func(_ view.Stat) bool { return true }),
		retry.WithEmitter(prom.NewEmitter()))
	require.Nil(t, err)

	require.Nil(t,
		r.Invoke(context.Background(), nil, nil,
			func(ctx context.Context, req, rsp interface{}) (err error) {
				return nil
			}))
}
