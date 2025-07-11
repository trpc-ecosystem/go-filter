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

package throttle

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenBucketInvalidArgs(t *testing.T) {
	_, err := NewTokenBucket(0, 0.1)
	require.NotNil(t, err)

	_, err = NewTokenBucket(-1, 0.1)
	require.NotNil(t, err)

	_, err = NewTokenBucket(MaximumTokens+1, 0.1)
	require.NotNil(t, err)

	_, err = NewTokenBucket(1, -0.1)
	require.NotNil(t, err)

	_, err = NewTokenBucket(1, 0.1)
	require.Nil(t, err)

	_, err = NewTokenBucket(10, 1)
	require.Nil(t, err)
}

func TestTokenBucket(t *testing.T) {
	tb, err := NewTokenBucket(2, 0.5)
	require.Nil(t, err)
	require.Equal(t, float64(1), tb.threshold)

	require.Equal(t, float64(2), math.Float64frombits(tb.tokens))
	require.True(t, tb.Allow())

	tb.OnFailure()
	require.Equal(t, float64(1), math.Float64frombits(tb.tokens))
	require.False(t, tb.Allow())
	tb.OnSuccess()
	require.Equal(t, 1.5, math.Float64frombits(tb.tokens))
	require.True(t, tb.Allow())

	tb.OnSuccess()
	tb.OnSuccess()
	require.Equal(t, float64(2), math.Float64frombits(tb.tokens))
	require.True(t, tb.Allow())

	tb.OnFailure()
	tb.OnFailure()
	tb.OnFailure()
	require.Equal(t, float64(0), math.Float64frombits(tb.tokens))
	require.False(t, tb.Allow())

	tb.OnSuccess()
	tb.OnSuccess()
	tb.OnSuccess()
	require.Equal(t, 1.5, math.Float64frombits(tb.tokens))
	require.True(t, tb.Allow())
}
