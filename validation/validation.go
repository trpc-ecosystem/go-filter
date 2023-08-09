// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package validation is the interceptor for parameter validation.
package validation

import (
	"context"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/log"
)

func init() {
	filter.Register(pluginName, ServerFilterWithOptions(defaultOptions), ClientFilterWithOptions(defaultOptions))
}

// defaultOptions is the default options of parameter.
var defaultOptions = options{
	LogFile:               nil,
	EnableErrorLog:        false,
	ServerValidateErrCode: errs.RetServerValidateFail,
	ClientValidateErrCode: errs.RetClientValidateFail,
}

// options is the options for parameter validation.
type options struct {
	LogFile               []bool `yaml:"logfile"`
	EnableErrorLog        bool   `yaml:"enable_error_log"`
	ServerValidateErrCode int    `yaml:"server_validate_err_code"`
	ClientValidateErrCode int    `yaml:"client_validate_err_code"`
}

// Option sets an option for the parameter.
type Option func(*options)

// Validator is the interface for automatic validation.
type Validator interface {
	Validate() error
}

// WithLogfile sets whether to log.
// Deprecated: Use WithErrorLog instead.
func WithLogfile(allow bool) Option {
	return func(opts *options) {
		opts.EnableErrorLog = allow
	}
}

// WithErrorLog sets whether to log errors.
func WithErrorLog(allow bool) Option {
	return func(opts *options) {
		opts.EnableErrorLog = allow
	}
}

// WithServerValidateErrCode sets the error code for server-side request validation failure.
func WithServerValidateErrCode(code int) Option {
	return func(opts *options) {
		opts.ServerValidateErrCode = code
	}
}

// WithClientValidateErrCode sets the error code for client-side response validation failure.
func WithClientValidateErrCode(code int) Option {
	return func(opts *options) {
		opts.ClientValidateErrCode = code
	}
}

// ServerFilter automatically validates the req input parameters during server-side RPC invocation.
// Deprecated: Use ServerFilterWithOptions instead.
func ServerFilter(opts ...Option) filter.ServerFilter {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}
	return ServerFilterWithOptions(o)
}

// ClientFilter automatically validates the rsp response parameters during client-side RPC invocation.
// Deprecated: Use ClientFilterWithOptions instead.
func ClientFilter(opts ...Option) filter.ClientFilter {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}
	return ClientFilterWithOptions(o)
}

// ServerFilterWithOptions automatically validates the req input parameters during server-side RPC invocation.
func ServerFilterWithOptions(o options) filter.ServerFilter {
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (interface{}, error) {
		// The request structure has not been validated by Validator.
		valid, ok := req.(Validator)
		if !ok {
			return handler(ctx, req)
		}

		// Verification passed.
		err := valid.Validate()
		if err == nil {
			return handler(ctx, req)
		}

		// Record logs as needed when the verification fails.
		errMsg := err.Error()
		if o.EnableErrorLog {
			if head, ok := ctx.Value(http.ContextKeyHeader).(*http.Header); ok {
				reqPath := head.Request.URL.Path
				reqRawQuery := head.Request.URL.RawQuery
				reqUserAgent := head.Request.Header.Get("User-Agent")
				reqReferer := head.Request.Header.Get("Referer")
				log.WithContext(ctx,
					log.Field{Key: "request_content", Value: req},
					log.Field{Key: "request_path", Value: reqPath},
					log.Field{Key: "request_query", Value: reqRawQuery},
					log.Field{Key: "request_useragent", Value: reqUserAgent},
					log.Field{Key: "request_referer", Value: reqReferer},
				).Errorf("validation request error: %s", errMsg)
			} else {
				log.WithContext(ctx, log.Field{Key: "request_content", Value: req}).
					Errorf("validation request error: %s", errMsg)
			}
		}

		return nil, errs.New(o.ServerValidateErrCode, errMsg)
	}
}

// ClientFilterWithOptions automatically validates the rsp response parameters during client-side RPC invocation.
func ClientFilterWithOptions(o options) filter.ClientFilter {
	return func(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
		// rsp does not need to be validated if An error occurred when calling downstream.
		if err := handler(ctx, req, rsp); err != nil {
			return err
		}

		// The response structure has not been validated by Validator.
		valid, ok := rsp.(Validator)
		if !ok {
			return nil
		}

		// Verification passed.
		err := valid.Validate()
		if err == nil {
			return nil
		}

		// Record logs as needed when the verification fails.
		if o.EnableErrorLog {
			log.WithContext(ctx, log.Field{Key: "response_content", Value: rsp}).
				Errorf("validation response error: %v", err)
		}

		return errs.New(o.ClientValidateErrCode, err.Error())
	}
}
