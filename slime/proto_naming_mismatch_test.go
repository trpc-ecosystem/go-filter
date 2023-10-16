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

package slime

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"trpc.group/trpc-go/trpc-go/filter"
	helloworld "trpc.group/trpc-go/trpc-go/testdata/trpc/helloworld"

	"github.com/stretchr/testify/require"
)

func TestCalleeServiceNameMismatch(t *testing.T) {
	f, err := ioutil.TempFile("", "trpc_go_*.yaml")
	require.Nil(t, err)
	defer func() {
		require.Nil(t, os.Remove(f.Name()))
	}()

	halfFailErr := errors.New("half fail")
	var n int
	filter.Register("half_fail", nil,
		func(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) error {
			n++
			if n%2 == 0 {
				return halfFailErr
			}
			return nil
		})

	_, err = f.WriteString(`
retry: &retry
  name: retry
  max_attempts: 2
  backoff:
    linear: [1ms]

client: &client
  filter:
    - slime
    - half_fail
  service:
    - callee: trpc.test.helloworld.Greeter
      name: this_is_different_from_callee
      retry_hedging:
        retry: *retry

plugins:
  slime:
    default: *client
`)
	require.Nil(t, err)

	require.Nil(t, trpc.LoadGlobalConfig(f.Name()))
	require.Nil(t, trpc.Setup(trpc.GlobalConfig()))

	require.Nil(t, SetRetryRetryableErr("retry", func(err error) bool {
		if err == halfFailErr {
			return true
		}
		return false
	}))

	c := helloworld.NewGreeterClientProxy()
	_, err = c.SayHello(context.Background(), &helloworld.HelloRequest{})
	require.Nil(t, err)
	_, err = c.SayHello(context.Background(), &helloworld.HelloRequest{})
	require.Nil(t, err)
}
