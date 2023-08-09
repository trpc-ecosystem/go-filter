// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"trpc.group/trpc-go/trpc-go/errs"
)

const retryBackoff = time.Millisecond * 50

func TestNewMaxAttempts(t *testing.T) {
	for maxAttempts, nilErr := range map[int]bool{
		-6:  false,
		-1:  false,
		0:   false,
		1:   true,
		3:   true,
		100: true,
	} {
		r, err := New(maxAttempts, []int{int(errs.RetClientNetErr)}, WithLinearBackoff(retryBackoff))
		require.Equal(t, nilErr, err == nil)
		if err == nil {
			require.LessOrEqual(t, r.maxAttempts, MaximumAttempts)
		}
	}
}

func TestNewBackoffPriority(t *testing.T) {
	r, err := New(
		3, []int{int(errs.RetClientNetErr)},
		WithBackoff(func(attempt int) time.Duration {
			return time.Second * time.Duration(attempt)
		}),
		WithExpBackoff(time.Millisecond, time.Millisecond*100, 2),
		WithLinearBackoff(time.Second))
	require.Nil(t, err)
	_, ok := r.bf.(*customizedBackoff)
	require.True(t, ok, "customized backoff should have the highest priority")

	r, err = New(
		3, []int{int(errs.RetClientNetErr)},
		WithExpBackoff(time.Millisecond, time.Millisecond*100, 2),
		WithLinearBackoff(time.Second))
	require.Nil(t, err)
	_, ok = r.bf.(*exponentialBackoff)
	require.True(t, ok, "exponential backoff should have higher priority than linear backoff")
}

func TestExpBackoff(t *testing.T) {
	exp, err := newExponentialBackoff(time.Millisecond, time.Second, 2)
	require.Nil(t, err)

	require.True(t, exp.backoff(1) <= time.Millisecond)
	require.True(t, exp.backoff(2) <= time.Millisecond*2)
	require.True(t, exp.backoff(3) <= time.Millisecond*4)
}

func TestLinearBackoff(t *testing.T) {
	_, err := newLinearBackoff()
	require.NotNil(t, err, "empty linear backoff should be invalid")

	l, err := newLinearBackoff(time.Millisecond, time.Millisecond*10)
	require.Nil(t, err)

	require.True(t, l.backoff(1) <= time.Millisecond)
	require.True(t, l.backoff(2) <= time.Millisecond*10)
	require.True(t, l.backoff(3) <= time.Millisecond*10)
}
