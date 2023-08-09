// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/filter"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

type Response struct {
	FirstMsg  string
	SecondMsg string
	ThirdMsg  string
}

type Handler func(http.ResponseWriter, *http.Request)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

func HTTPListenAndServe(h Handler) (target string, stop func()) {
	errCh := make(chan error, 1)
	for {
		port := rand.Int()%64511 + 1024
		s := http.Server{Addr: fmt.Sprintf(":%d", port), Handler: h}
		go func() {
			errCh <- s.ListenAndServe()
		}()
		select {
		case <-time.After(time.Millisecond * 500):
			// sleep a while to wait server start.
			return fmt.Sprintf("ip://localhost:%d", port), func() {
				_ = s.Close()
			}
		case err := <-errCh:
			log.Printf("failed to ListenAndServe, err: %s", err)
		}
	}
}

var (
	cases = []func(c thttp.Client, rsp interface{}) error{
		func(c thttp.Client, rsp interface{}) error {
			return c.Get(trpc.BackgroundContext(), "/any", rsp)
		},
		func(c thttp.Client, rsp interface{}) error {
			return c.Post(trpc.BackgroundContext(), "/any", nil, rsp)
		},
		func(c thttp.Client, rsp interface{}) error {
			return c.Put(trpc.BackgroundContext(), "/any", nil, rsp)
		},
		func(c thttp.Client, rsp interface{}) error {
			return c.Delete(trpc.BackgroundContext(), "/any", nil, rsp)
		},
	}
)

func testCommon(
	t *testing.T,
	intercept filter.ClientFilter,
	run func(c thttp.Client, rsp interface{}) error,
) {
	var calledN atomic.Int32
	target, stop := HTTPListenAndServe(
		func(w http.ResponseWriter, r *http.Request) {
			if calledN.Inc() == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write(jsonMarshal(Response{FirstMsg: "first"}))
			} else if calledN.Load() == 2 {
				_, _ = w.Write(jsonMarshal(Response{SecondMsg: "second"}))
			} else {
				_, _ = w.Write(jsonMarshal(Response{ThirdMsg: "third"}))
			}
		},
	)
	defer stop()

	var rsp Response
	err := run(
		thttp.NewClientProxy("",
			client.WithTarget(target),
			client.WithFilter(intercept)),
		&rsp)
	require.Nil(t, err)
	require.Equal(t, int32(3), calledN.Load())
	require.Empty(t, rsp.FirstMsg, "FirstMsg should not be polluted")
	require.Empty(t, rsp.SecondMsg, "SecondMsg should not be polluted")
	require.Equal(t, "third", rsp.ThirdMsg)
}

func jsonMarshal(v interface{}) []byte {
	bts, _ := json.Marshal(v)
	return bts
}

func testHTTPResponseBodyOnError(
	t *testing.T,
	retryOrHedging filter.ClientFilter,
) {
	target, stop := HTTPListenAndServe(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("always fail"))
	})
	defer stop()

	var httpRsp *http.Response
	readHTTPRspBody := func(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) (err error) {
		err = f(ctx, req, rsp)
		msg := codec.Message(ctx)
		httpRsp = msg.ClientRspHead().(*thttp.ClientRspHeader).Response
		return
	}

	c := thttp.NewClientProxy("",
		client.WithTarget(target),
		client.WithFilters([]filter.ClientFilter{readHTTPRspBody, retryOrHedging}))

	err := c.Get(trpc.BackgroundContext(), "/", nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, httpRsp.StatusCode)
	body, err := ioutil.ReadAll(httpRsp.Body)
	require.Nil(t, err)
	require.Equal(t, "always fail", string(body))
}
