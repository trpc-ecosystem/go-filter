// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package debuglog is a logger trpc-filter to printing server/client RPC calls.
package debuglog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/log"
)

func init() {
	filter.Register("debuglog", ServerFilter(), ClientFilter())
	filter.Register(
		"simpledebuglog", ServerFilter(WithLogFunc(SimpleLogFunc)),
		ClientFilter(WithLogFunc(SimpleLogFunc)),
	)
	filter.Register(
		"pjsondebuglog", ServerFilter(WithLogFunc(PrettyJSONLogFunc)),
		ClientFilter(WithLogFunc(PrettyJSONLogFunc)),
	)
	filter.Register(
		"jsondebuglog", ServerFilter(WithLogFunc(JSONLogFunc)),
		ClientFilter(WithLogFunc(JSONLogFunc)),
	)
}

// options are the configuration options.
type options struct {
	logFunc         LogFunc
	errLogLevelFunc LogLevelFunc
	nilLogLevelFunc LogLevelFunc
	enableColor     bool
	include         []*RuleItem
	exclude         []*RuleItem
}

// passed is the filtering result.
// When it is true, will go to the logging process.
func (o *options) passed(rpcName string, errCode int) bool {
	// Calculation of the include rule.
	for _, in := range o.include {
		if in.Matched(rpcName, errCode) {
			return true
		}
	}
	// If the include rule is configured, the exclude rule will not be matched.
	if len(o.include) > 0 {
		return false
	}

	// Calculation of the exclude rule.
	for _, ex := range o.exclude {
		if ex.Matched(rpcName, errCode) {
			return false
		}
	}
	return true
}

// Option sets the optiopns.
type Option func(*options)

// LogFunc is the struct print method function.
type LogFunc func(ctx context.Context, req, rsp interface{}) string

// LogLevelFunc specifies the log level.
type LogLevelFunc func(ctx context.Context, format string, args ...interface{})

// WithLogFunc sets the print body method.
func WithLogFunc(f LogFunc) Option {
	return func(opts *options) {
		opts.logFunc = f
	}
}

// WithErrLogLevelFunc sets the log level print method.
func WithErrLogLevelFunc(f LogLevelFunc) Option {
	return func(opts *options) {
		opts.errLogLevelFunc = f
	}
}

// WithNilLogLevelFunc sets the non-error log level print method.
func WithNilLogLevelFunc(f LogLevelFunc) Option {
	return func(opts *options) {
		opts.nilLogLevelFunc = f
	}
}

// WithInclude sets the include options.
func WithInclude(in *RuleItem) Option {
	return func(opts *options) {
		opts.include = append(opts.include, in)
	}
}

// WithExclude sets the exclude options.
func WithExclude(ex *RuleItem) Option {
	return func(opts *options) {
		opts.exclude = append(opts.exclude, ex)
	}
}

// WithEnableColor enable multiple color log output.
func WithEnableColor(enable bool) Option {
	return func(opts *options) {
		opts.enableColor = enable
	}
}

// logLevel is log level.
type logLevel = string

var (
	traceLevel   logLevel = "trace"
	debugLevel   logLevel = "debug"
	warningLevel logLevel = "warning"
	infoLevel    logLevel = "info"
	errorLevel   logLevel = "error"
	fatalLevel   logLevel = "fatal"
)

// LogContextfFuncs is a map of methods for logging at different levels.
var LogContextfFuncs = map[string]func(ctx context.Context, format string, args ...interface{}){
	traceLevel:   log.TraceContextf,
	debugLevel:   log.DebugContextf,
	warningLevel: log.WarnContextf,
	infoLevel:    log.InfoContextf,
	errorLevel:   log.ErrorContextf,
	fatalLevel:   log.FatalContextf,
}

// DefaultLogFunc is the default struct print method.
var DefaultLogFunc = func(ctx context.Context, req, rsp interface{}) string {
	return fmt.Sprintf(", req:%+v, rsp:%+v", req, rsp)
}

// SimpleLogFunc does not print the struct.
var SimpleLogFunc = func(ctx context.Context, req, rsp interface{}) string {
	return ""
}

// PrettyJSONLogFunc is the method for printing formatted JSON.
var PrettyJSONLogFunc = func(ctx context.Context, req, rsp interface{}) string {
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	rspJSON, _ := json.MarshalIndent(rsp, "", "  ")
	return fmt.Sprintf("\nreq:%s\nrsp:%s", string(reqJSON), string(rspJSON))
}

// JSONLogFunc is the method for printing JSON.
var JSONLogFunc = func(ctx context.Context, req, rsp interface{}) string {
	reqJSON, _ := json.Marshal(req)
	rspJSON, _ := json.Marshal(rsp)
	return fmt.Sprintf("\nreq:%s\nrsp:%s", string(reqJSON), string(rspJSON))
}

// ServerFilter is the server-side filter.
func ServerFilter(opts ...Option) filter.ServerFilter {
	o := getFilterOptions(opts...)
	nilLogFormat := getLogFormat(debugLevel, o.enableColor, "server request:%s, cost:%s, from:%s%s")
	errLogFormat := getLogFormat(errorLevel, o.enableColor, "server request:%s, cost:%s, from:%s, err:%s%s")
	deadlineLogFormat := getLogFormat(errorLevel, o.enableColor,
		"server request:%s, cost:%s, from:%s, err:%s, total timeout:%s%s")
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (rsp interface{}, err error) {
		begin := time.Now()
		rsp, err = handler(ctx, req)
		msg := trpc.Message(ctx)
		if !o.passed(msg.ServerRPCName(), errs.Code(err)) {
			return rsp, err
		}

		end := time.Now()
		var addr string
		if msg.RemoteAddr() != nil {
			addr = msg.RemoteAddr().String()
		}
		if err == nil {
			o.nilLogLevelFunc(
				ctx, nilLogFormat,
				msg.ServerRPCName(), end.Sub(begin), addr, o.logFunc(ctx, req, rsp),
			)
		} else {
			deadline, ok := ctx.Deadline()
			if ok {
				o.errLogLevelFunc(
					ctx, deadlineLogFormat, msg.ServerRPCName(), end.Sub(begin), addr, err.Error(),
					deadline.Sub(begin), o.logFunc(ctx, req, rsp),
				)
			} else {
				o.errLogLevelFunc(
					ctx, errLogFormat, msg.ServerRPCName(), end.Sub(begin), addr, err.Error(), o.logFunc(ctx, req, rsp),
				)
			}
		}
		return rsp, err
	}
}

// ClientFilter is the client-side filter.
func ClientFilter(opts ...Option) filter.ClientFilter {
	o := getFilterOptions(opts...)
	nilLogFormat := getLogFormat(debugLevel, o.enableColor, "client request:%s, cost:%s, to:%s%s")
	errLogFormat := getLogFormat(errorLevel, o.enableColor, "client request:%s, cost:%s, to:%s, err:%s%s")
	return func(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) (err error) {
		msg := trpc.Message(ctx)
		begin := time.Now()
		err = handler(ctx, req, rsp)
		if !o.passed(msg.ClientRPCName(), errs.Code(err)) {
			return err
		}

		end := time.Now()
		var addr string
		if msg.RemoteAddr() != nil {
			addr = msg.RemoteAddr().String()
		}
		if err == nil {
			o.nilLogLevelFunc(
				ctx, nilLogFormat, msg.ClientRPCName(), end.Sub(begin), addr, o.logFunc(ctx, req, rsp),
			)
		} else {
			o.errLogLevelFunc(
				ctx, errLogFormat, msg.ClientRPCName(), end.Sub(begin), addr, err.Error(), o.logFunc(ctx, req, rsp),
			)
		}
		return err
	}
}

// getFilterOptions gets the interceptor condition options.
func getFilterOptions(opts ...Option) *options {
	o := &options{
		logFunc:         DefaultLogFunc,
		errLogLevelFunc: LogContextfFuncs[errorLevel],
		nilLogLevelFunc: LogContextfFuncs[debugLevel],
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// getLogLevelFunc gets the log print method for the corresponding log level.
func getLogLevelFunc(level string, defaultLevel string) LogLevelFunc {
	logFunc, ok := LogContextfFuncs[level]
	if !ok {
		logFunc = LogContextfFuncs[defaultLevel]
	}
	return logFunc
}

// color is the type for log color.
type color uint8

// fontColor returns the font color of the log.
func (c color) fontColor() uint8 {
	return uint8(c)
}

const (
	colorRed     color = 31 // Red color.
	colorMagenta color = 35 // Magenta color.
)

// levelColorMap maps log levels to colors.
var levelColorMap = map[logLevel]color{
	debugLevel: colorMagenta,
	errorLevel: colorRed,
}

// getLogFormat formats the log output with different colors based on user settings.
func getLogFormat(level logLevel, enableColor bool, format string) string {
	if !enableColor {
		// Default scheme.
		return format
	}
	preColor := "\033[1;%dm" // Control character for color display.
	sufColor := "\033[0m"    // Color display end character.
	if v, ok := levelColorMap[level]; ok {
		return fmt.Sprintf(preColor, v.fontColor()) + format + sufColor
	}
	return format
}
