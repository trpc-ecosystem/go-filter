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

package http_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-filter/slime/hedging"
)

const (
	hedgingDelay = time.Millisecond * 10
)

func TestHedgingHTTP(t *testing.T) {
	h, err := hedging.New(
		3, []int{},
		hedging.WithNonFatalErr(func(err error) bool { return true }),
		hedging.WithStaticHedgingDelay(hedgingDelay),
		hedging.WithRspToErr(func(rsp interface{}) error {
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
		testCommon(t, h.Invoke, cs)
	}
}

func TestHedgingHTTP_ResponseBodyOnError(t *testing.T) {
	h, err := hedging.New(
		2, []int{},
		hedging.WithNonFatalErr(func(err error) bool { return true }),
		hedging.WithStaticHedgingDelay(hedgingDelay))
	require.Nil(t, err)

	testHTTPResponseBodyOnError(t, h.Invoke)
}
