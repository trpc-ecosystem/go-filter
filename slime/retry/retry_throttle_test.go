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

package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"

	"trpc.group/trpc-go/trpc-filter/slime/pushback"
	"trpc.group/trpc-go/trpc-filter/slime/retry"
)

type records struct {
	successN, failureN int
	history            []bool
}

func (r *records) OnSuccess() {
	r.successN++
	r.history = append(r.history, true)
}

func (r *records) OnFailure() {
	r.failureN++
	r.history = append(r.history, false)
}

type throttleAlwaysDisallow struct {
	records
}

func (t *throttleAlwaysDisallow) Allow() bool { return false }

type throttleAlwaysAllow struct {
	records
}

func (t *throttleAlwaysAllow) Allow() bool { return true }

func TestThrottleNeverThrottleFirstTry(t *testing.T) {
	throttle := &throttleAlwaysDisallow{}
	tr, err := retry.NewThrottledRetry(
		3, []int{int(errs.RetClientNetErr)},
		throttle,
		WithLog(),
		retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Add(1)
		return errs.New(errs.RetClientNetErr, "")
	}

	err = tr.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(1), attempts.Load())
	require.Equal(t, 1, throttle.failureN)
}

// If server pushback issues no hedging, throttle always calls OnFailure,
// regardless of whether the request is successful or not.
func TestThrottleServerPushbackNoHedging(t *testing.T) {
	for _, retErr := range []error{
		nil,
		errs.New(errs.RetClientNetErr, ""),
		errors.New("fatal error"),
	} {
		h, err := retry.New(
			3, []int{int(errs.RetClientNetErr)},
			WithLog(),
			retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
		require.Nil(t, err)

		throttle := &throttleAlwaysAllow{}
		tr := h.NewThrottledRetry(throttle)

		var attempts atomic.Int32
		handler := func(ctx context.Context, req, rsp interface{}) error {
			attempts.Add(1)
			codec.Message(ctx).ClientMetaData()[pushback.MetaKey] = []byte(time.Duration(-1).String())
			return retErr
		}

		err = tr.Invoke(context.Background(), nil, nil, handler)
		require.Equal(t, retErr, err)
		require.Equal(t, int32(1), attempts.Load())
		require.Equal(t, 1, throttle.failureN,
			"request should be always counted as failure if server pushback issues no hedging")
	}
}

// Fatal error does not take count into throttle.
func TestThrottleFatalError(t *testing.T) {
	h, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	throttle := &throttleAlwaysAllow{}
	tr := h.NewThrottledRetry(throttle)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Add(1)
		return errors.New("fatal error")
	}

	err = tr.Invoke(context.Background(), nil, nil, handler)
	require.NotNil(t, err)
	require.Equal(t, int32(1), attempts.Load())
	require.Equal(t, 0, throttle.failureN,
		"throttle should not count fatal error")
	require.Equal(t, 0, throttle.successN)
	require.Equal(t, 0, len(throttle.history))
}

func TestThrottleNonFatalError(t *testing.T) {
	h, err := retry.New(
		3, []int{int(errs.RetClientNetErr)},
		WithLog(),
		retry.WithBackoff(func(int) time.Duration { return retryBackoff }))
	require.Nil(t, err)

	throttle := &throttleAlwaysAllow{}
	tr := h.NewThrottledRetry(throttle)

	var attempts atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempts.Add(1)
		if attempts.Load() == 3 {
			return nil
		}
		return errs.New(errs.RetClientNetErr, "")
	}

	err = tr.Invoke(context.Background(), nil, nil, handler)
	require.Nil(t, err)
	require.Equal(t, int32(3), attempts.Load())
	require.Equal(t, 2, throttle.failureN)
	require.Equal(t, 1, throttle.successN)
}
