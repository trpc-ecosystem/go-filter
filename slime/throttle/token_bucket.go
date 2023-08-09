// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package throttle implement the throttle for retry/hedging request.
package throttle

import (
	"fmt"
	"math"
	"sync/atomic"
)

const (
	// MaximumTokens defines the up limit of tokens for TokenBucket.
	MaximumTokens = 1000
)

// TokenBucket defines a throttle based on token bucket.
// TokenBucket's methods may be called by multiple goroutines simultaneously.
type TokenBucket struct {
	tokens     uint64
	maxTokens  float64
	threshold  float64
	tokenRatio float64
}

// NewTokenBucket create a new TokenBucket.
func NewTokenBucket(maxTokens, tokenRatio float64) (*TokenBucket, error) {
	if maxTokens > MaximumTokens {
		return nil, fmt.Errorf("expect tokens less or equal to %d, got %f", MaximumTokens, maxTokens)
	}

	if maxTokens <= 0 {
		return nil, fmt.Errorf("expect positive tokens, got %f", maxTokens)
	}

	if tokenRatio <= 0 {
		return nil, fmt.Errorf("expect positive taken ratio, got %f", tokenRatio)
	}

	return &TokenBucket{
		tokens:     math.Float64bits(maxTokens),
		maxTokens:  maxTokens,
		threshold:  maxTokens / 2,
		tokenRatio: tokenRatio,
	}, nil
}

// Allow whether a new request could be issued.
//
//go:nosplit
func (tb *TokenBucket) Allow() bool {
	return math.Float64frombits(atomic.LoadUint64(&tb.tokens)) > tb.threshold
}

// OnSuccess increase tokens in bucket by token ratio, but not greater than maxTokens.
//
//go:nosplit
func (tb *TokenBucket) OnSuccess() {
	for {
		tokens := math.Float64frombits(atomic.LoadUint64(&tb.tokens))
		if tokens == tb.maxTokens {
			return
		}

		newTokens := tokens + tb.tokenRatio
		if newTokens > tb.maxTokens {
			newTokens = tb.maxTokens
		}

		if atomic.CompareAndSwapUint64(
			&tb.tokens,
			math.Float64bits(tokens),
			math.Float64bits(newTokens),
		) {
			break
		}
	}
}

// OnFailure decrease tokens in bucket by 1, but not less than 0.
//
//go:nosplit
func (tb *TokenBucket) OnFailure() {
	for {
		tokens := math.Float64frombits(atomic.LoadUint64(&tb.tokens))
		if tokens == 0 {
			return
		}

		newTokens := tokens - 1
		if newTokens < 0 {
			newTokens = 0
		}

		if atomic.CompareAndSwapUint64(
			&tb.tokens,
			math.Float64bits(tokens),
			math.Float64bits(newTokens),
		) {
			break
		}
	}
}
