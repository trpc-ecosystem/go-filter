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

package jwt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	thttp "trpc.group/trpc-go/trpc-go/http"

	"github.com/stretchr/testify/assert"
)

// fakeServerHandleFunc empty handle for test
func fakeServerHandleFunc(ctx context.Context, req interface{}) (rsp interface{}, err error) {
	return rsp, nil
}

func mockHttpRequest() *http.Request {
	sign := mockSigner()
	SetDefaultSigner(sign)
	token, _ := sign.Sign(mockUserInfo())
	req := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/v1"}, Header: http.Header{}}
	req.Header.Add("Authorization", "Bearer "+token)
	return req
}

func TestServerFilter(t *testing.T) {
	f := ServerFilter(WithExcludePathSet(map[string]bool{"/v1/login": true}))
	r := mockHttpRequest()
	w := &httptest.ResponseRecorder{}
	m := &thttp.Header{Request: r, Response: w}
	ctx := thttp.WithHeader(context.Background(), m)
	_, err := f(ctx, []byte("req"), fakeServerHandleFunc)
	assert.Nil(t, err)
}
