// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package retry

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// linearBackoff has lowest priority in all kinds of backoff.
type linearBackoff struct {
	bfs []time.Duration
}

// newLinearBackoff create a new linear backoff. Empty bfs will cause an error.
func newLinearBackoff(bfs ...time.Duration) (*linearBackoff, error) {
	if len(bfs) == 0 {
		return nil, errors.New("linear backoff list must not be empty")
	}
	return &linearBackoff{bfs: bfs}, nil
}

// backoff is randomly distributed in [0, bfs[min(len(bfs)-1, attempt)]].
func (bf *linearBackoff) backoff(attempt int) (delay time.Duration) {
	defer func() {
		delay = time.Duration(rand.Float64() * float64(delay))
	}()

	if attempt <= 0 {
		return 0
	}

	l := len(bf.bfs)
	if attempt <= l {
		return bf.bfs[attempt-1]
	}
	return bf.bfs[l-1]
}
