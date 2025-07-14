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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"trpc.group/trpc-go/trpc-go/errs"
)

const hedgingDelay = time.Millisecond * 50

func TestNewMaxAttempts(t *testing.T) {
	for maxAttempts, nilErr := range map[int]bool{
		-6:  false,
		-1:  false,
		0:   false,
		1:   true,
		3:   true,
		100: true,
	} {
		h, err := New(maxAttempts, []int{int(errs.RetClientNetErr)}, WithStaticHedgingDelay(hedgingDelay))
		require.Equal(t, nilErr, err == nil)
		if err == nil {
			require.LessOrEqual(t, h.maxAttempts, MaximumAttempts)
		}
	}
}
