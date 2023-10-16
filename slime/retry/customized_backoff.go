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
	"time"
)

// customizedBackoff wraps an user defined bf function.
type customizedBackoff struct {
	bf func(attempt int) time.Duration
}

// newCustomizedBackoff create a new customizedBackoff.
func newCustomizedBackoff(bf func(attempt int) time.Duration) (*customizedBackoff, error) {
	if bf == nil {
		return nil, errors.New("provided bf function must not be nil")
	}

	return &customizedBackoff{bf: bf}, nil
}

// backoff simply calls the wrapped bf function.
func (bf *customizedBackoff) backoff(attempt int) time.Duration {
	return bf.bf(attempt)
}
