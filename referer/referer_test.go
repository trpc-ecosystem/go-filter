// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package referer

import (
	"context"
	stdhttp "net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/http"
)

// TestPlugin_Type TestPlugin_Type
func TestPlugin_Type(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, pluginType, p.Type())
}

func Test_matchedReferer(t *testing.T) {
	matchedReferer([]string{"test.qq.com"}, "")
}

func buildHeader(ctx context.Context, referer string, urlStr string) context.Context {
	header := make(stdhttp.Header)
	header["Referer"] = []string{referer}
	urlReq, _ := url.Parse(urlStr)
	h := &http.Header{Request: &stdhttp.Request{URL: urlReq, Header: header}}
	ctx = context.WithValue(ctx, http.ContextKeyHeader, h)
	return ctx
}

func TestServerFilter(t *testing.T) {
	f := ServerFilter(WithRefererDomain("/test/url", "test.qq.com"))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return &struct{}{}, nil }
	ctx := buildHeader(trpc.BackgroundContext(), "https://test.qq.com/demo.html", "https://test.qq.com/test/url")
	req := &codec.Body{}
	rsp, err := f(ctx, req, handler)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)

	ctx = buildHeader(trpc.BackgroundContext(), "https://test.1qq.com/demo.html", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.NotNil(t, err)
	assert.Nil(t, rsp)

	f = ServerFilter(WithRefererDomain(refererApplyAllPath, "test.qq.com"))
	ctx = buildHeader(trpc.BackgroundContext(), "https://test.qq.com/demo.html", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)

	f = ServerFilter(WithRefererDomain(refererApplyAllPath))
	ctx = buildHeader(trpc.BackgroundContext(), "https://test.qq.com/demo.html", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.NotNil(t, err)
	assert.Nil(t, rsp)

	f = ServerFilter(WithRefererDomain(refererApplyAllPath, "NULL"))
	ctx = buildHeader(trpc.BackgroundContext(), "", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)

	f = ServerFilter(WithRefererDomain(refererApplyAllPath, "*"))
	ctx = buildHeader(trpc.BackgroundContext(), "https://test.ddddqq.com/demo.html", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)

	f = ServerFilter(WithRefererDomain("/test/url", "test.qq.com"))
	ctx = buildHeader(trpc.BackgroundContext(), "https://test.ddddqq.com/demo.html", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.NotNil(t, err)
	assert.Nil(t, rsp)

	f = ServerFilter(WithRefererDomain("/test/url", ".qq.com"))
	ctx = buildHeader(trpc.BackgroundContext(), "https://test.qq.com/demo.html", "https://test.qq.com/test/url")
	rsp, err = f(ctx, req, handler)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
