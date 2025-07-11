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

package once_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"trpc.group/trpc-go/trpc-filter/slime/once"
)

func TestInvoke(t *testing.T) {
	o := once.New()

	var attempt atomic.Int32
	handler := func(ctx context.Context, req, rsp interface{}) error {
		attempt.Inc()
		*rsp.(*int) = 1
		return nil
	}

	var rsp int
	err := o.Invoke(context.Background(), nil, &rsp, handler)
	require.Nil(t, err)
	require.Equal(t, rsp, 1)
	require.Equal(t, int32(1), attempt.Load())
}
