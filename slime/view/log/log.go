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

// Package log provides log utils used by retry/hedging.
package log

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	timeFormat = "15:04:05.000"
)

// Logger is a simple interface to print an string.
type Logger interface {
	Println(string)
}

// CtxLogger accepts an additional context.
type CtxLogger interface {
	Println(context.Context, string)
}

// LazyLog buffers messages and flush them by Logger.
//
// LazyLog is not concurrent safe.
type LazyLog struct {
	log CtxLogger
	buf []string
}

// NewLazyLog create a new LazyLog.
func NewLazyLog(log Logger) *LazyLog {
	return &LazyLog{log: &noopCtxLog{log: log}, buf: []string{"[lazy log]"}}
}

// NewLazyCtxLog creates a new LazyLog.
func NewLazyCtxLog(log CtxLogger) *LazyLog {
	return &LazyLog{log: log, buf: []string{"[lazy log]"}}
}

// Printf provides a format printer.
func (l *LazyLog) Printf(format string, a ...interface{}) {
	l.buf = append(l.buf, time.Now().Format(timeFormat)+"]\t"+fmt.Sprintf(format, a...))
}

// Flush flushes messages in buffer. Messages are separated by a new line
// and flushed with a single call of Logger.Println.
func (l *LazyLog) Flush() {
	l.FlushCtx(context.Background())
}

// FlushCtx flushes messages in buffer. Messages are separated by a new line
// and flushed with a single call of Logger.Println.
func (l *LazyLog) FlushCtx(ctx context.Context) {
	l.log.Println(ctx, strings.Join(l.buf, "\n"))
	l.buf = l.buf[:0]
}

type noopCtxLog struct {
	log Logger
}

func (l *noopCtxLog) Println(_ context.Context, s string) {
	l.log.Println(s)
}
