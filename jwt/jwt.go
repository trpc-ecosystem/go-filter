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

// Package jwt 身份认证
package jwt

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	trpcHttp "trpc.group/trpc-go/trpc-go/http"
)

// ContextKey 定义类型
type ContextKey string

const (
	// AuthJwtCtxKey context key
	AuthJwtCtxKey = ContextKey("AuthJwtCtxKey")
)

// DefaultSigner 默认的 signer
var DefaultSigner Signer

// SetDefaultSigner 设置默认 signer
func SetDefaultSigner(s Signer) {
	if s != nil {
		DefaultSigner = s
	}
}

// DefaultParseTokenFunc 默认获取 token 的函数, 用户可自定义实现
var DefaultParseTokenFunc = func(ctx context.Context, req interface{}) (string, error) {
	head := trpcHttp.Head(ctx)
	token := head.Request.Header.Get("Authorization")
	return strings.TrimPrefix(token, "Bearer "), nil
}

// options 插件配置
type options struct {
	ExcludePathSet map[string]bool // path 白名单
}

// isInExcludePath 是否在白名单列表中
func (o *options) isInExcludePath(path string) bool {
	_, ok := o.ExcludePathSet[path]
	return ok
}

// Option 设置参数选项
type Option func(*options)

// WithExcludePathSet 设置 path 白名单
func WithExcludePathSet(set map[string]bool) Option {
	return func(o *options) {
		o.ExcludePathSet = set
	}
}

// ServerFilter 设置服务端增加 jwt 验证
func ServerFilter(opts ...Option) filter.ServerFilter {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (interface{}, error) {
		var head = trpcHttp.Head(ctx)
		// 非http请求
		if head == nil {
			return handler(ctx, req)
		}
		// 是否跳过OA验证(path白名单)
		var r = head.Request
		if r.URL != nil && o.isInExcludePath(r.URL.Path) {
			return handler(ctx, req)
		}
		token, err := DefaultParseTokenFunc(ctx, req)
		if err != nil {
			return nil, err
		}
		customInfo, err := DefaultSigner.Verify(token)
		if err != nil {
			return nil, errs.NewFrameError(errs.RetServerAuthFail, err.Error())
		}
		// 将认证信息存储到ctx中用于业务侧使用
		innerCtx := context.WithValue(ctx, AuthJwtCtxKey, customInfo)
		return handler(innerCtx, req)
	}
}

// GetCustomInfo 获取用户信息 (参数 ptr 为struct的指针对象)
func GetCustomInfo(ctx context.Context, ptr interface{}) error {
	if data, ok := ctx.Value(AuthJwtCtxKey).(map[string]interface{}); ok {
		return mapstructure.Decode(data, ptr)
	}
	return fmt.Errorf("fail to find ctx value! key=(%v)", AuthJwtCtxKey)
}
