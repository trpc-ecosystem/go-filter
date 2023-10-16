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
	"errors"
	"math"
	"math/rand"
	"time"
)

// exponentialBackoff has a priority higher than linearBackoff but lower than customizedBackoff.
type exponentialBackoff struct {
	initial    time.Duration
	maximum    time.Duration
	multiplier int
}

// newExponentialBackoff create a new exponentialBackoff.
func newExponentialBackoff(initial, maximum time.Duration, multiplier int) (*exponentialBackoff, error) {
	if initial <= 0 {
		return nil, errors.New("initial of exponential backoff must be positive")
	}

	if maximum < initial {
		return nil, errors.New("maximum of exponential backoff must be greater than initial")
	}

	if multiplier <= 0 {
		return nil, errors.New("multiplier of exponential backoff must be positive")
	}

	return &exponentialBackoff{
		initial:    initial,
		maximum:    maximum,
		multiplier: multiplier,
	}, nil
}

// backoff returns the backoff after each attempt. The result is randomized.
func (bf *exponentialBackoff) backoff(attempt int) time.Duration {
	ceil := math.Min(
		float64(bf.initial)*math.Pow(
			float64(bf.multiplier),
			float64(attempt-1)),
		float64(bf.maximum))
	return time.Duration(rand.Float64() * ceil)
}
