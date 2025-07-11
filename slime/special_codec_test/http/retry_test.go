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

package http_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-filter/slime/retry"
)

const (
	retryBackoff = time.Millisecond * 10
)

func TestRetryHTTP(t *testing.T) {
	r, err := retry.New(
		3, []int{},
		retry.WithRetryableErr(func(error) bool { return true }),
		retry.WithLinearBackoff(retryBackoff),
		retry.WithRspToErr(func(rsp interface{}) error {
			switch v := rsp.(type) {
			case *Response:
				if v.SecondMsg == "second" {
					return fmt.Errorf("rsp should retry")
				}
			default:
				return nil
			}
			return nil
		}))
	require.Nil(t, err)
	for _, cs := range cases {
		testCommon(t, r.Invoke, cs)
	}
}

func TestRetryHTTP_ResponseBodyOnError(t *testing.T) {
	h, err := retry.New(
		2, []int{},
		retry.WithRetryableErr(func(err error) bool { return true }),
		retry.WithLinearBackoff(retryBackoff))
	require.Nil(t, err)

	testHTTPResponseBodyOnError(t, h.Invoke)
}
