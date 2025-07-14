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

// Package referer 安全验证
package referer

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/http"
)

const (
	refererPrefixHTTP   = "http://"
	refererPrefixHTTPS  = "https://"
	refererApplyAllPath = "apply_to_all_path"
)

// DefRefererErrorFunc 默认错误处理器
var DefRefererErrorFunc = func(ctx context.Context, referer string, err error) error {
	return err
}

func init() {
	filter.Register(pluginName, ServerFilter(), nil)
}

// options 参数选项
type options struct {
	AllowReferer map[string][]string
}

// Option 设置参数选项
type Option func(*options)

// WithRefererDomain 设置每个接口对应允许Referer域名
func WithRefererDomain(rpcName string, domains ...string) Option {
	return func(opts *options) {
		if opts.AllowReferer == nil {
			opts.AllowReferer = make(map[string][]string)
		}
		domain, ok := opts.AllowReferer[rpcName]
		if !ok {
			domain = make([]string, 0)
		}
		domain = append(domain, domains...)
		opts.AllowReferer[rpcName] = domain
	}
}

// ServerFilter 设置服务端增加Referer验证
func ServerFilter(opts ...Option) filter.ServerFilter {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (rsp interface{}, err error) {
		head, ok := ctx.Value(http.ContextKeyHeader).(*http.Header)
		if !ok {
			return handler(ctx, req)
		}

		referer := head.Request.Header.Get("Referer")
		if o.AllowReferer == nil {
			return nil, DefRefererErrorFunc(ctx, referer, errs.NewFrameError(errs.RetServerAuthFail,
				"allowReferer config empty"))
		}

		var refererList []string
		if list, ok := o.AllowReferer[head.Request.URL.Path]; ok {
			refererList = list
		} else if list, ok := o.AllowReferer[refererApplyAllPath]; ok { // 支持全局配置
			refererList = list
		} else {
			return nil, DefRefererErrorFunc(ctx, referer, errs.NewFrameError(errs.RetServerAuthFail,
				fmt.Sprintf("this url does not allow access from %s", referer)))
		}

		parsedOriginHost := ""
		if referer != "" {
			if !(strings.HasPrefix(referer, refererPrefixHTTP) ||
				strings.HasPrefix(referer, refererPrefixHTTPS)) {
				return nil, DefRefererErrorFunc(ctx, referer, errs.NewFrameError(errs.RetServerAuthFail,
					fmt.Sprintf("referer %s prefix err !", referer)))
			}

			parsedOriginObj, err := url.Parse(referer)
			if err != nil {
				return nil, DefRefererErrorFunc(ctx, referer, err)
			}
			parsedOriginHost = parsedOriginObj.Host
		}

		matched := matchedReferer(refererList, parsedOriginHost)
		if !matched {
			return nil, DefRefererErrorFunc(ctx, referer, errs.NewFrameError(errs.RetServerAuthFail,
				fmt.Sprintf("this url does not allow access from %s", referer)))
		}

		// 鉴权成功,执行后续流程
		return handler(ctx, req)
	}
}

func matchedReferer(refererList []string, parsedOriginHost string) bool {
	for _, domain := range refererList {
		// 当referer 不为空时候 允许 * 配置
		if parsedOriginHost != "" {
			if domain == "*" {
				return true
			}
			if strings.HasSuffix(parsedOriginHost, fixDomain(domain)) {
				return true
			}
			if parsedOriginHost == domain {
				return true
			}
		} else if domain == "NULL" {
			// 当不存在referer时候 只允许配置了NULL 通过
			return true
		}
	}
	return false
}

func fixDomain(domain string) string {
	if !strings.HasPrefix(domain, ".") {
		return "." + domain
	}
	return domain
}
