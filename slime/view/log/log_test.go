// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-filter/slime/view/log"
)

type bufLog struct {
	buf string
}

func (l *bufLog) Println(s string) {
	l.buf += s
}

type ctxBufLog struct {
	buf string
}

func (l *ctxBufLog) Println(ctx context.Context, s string) {
	l.buf += s
}

func TestLazyLog(t *testing.T) {
	bufLog := bufLog{}
	l := log.NewLazyLog(&bufLog)
	l.Printf("aaa")
	l.Printf("%s", "bbb")
	require.Equal(t, "", bufLog.buf)

	l.Flush()
	require.Contains(t, bufLog.buf, "aaa")
	require.Contains(t, bufLog.buf, "bbb")

	l.Printf("ccc")
	l.Flush()
	require.Contains(t, bufLog.buf, "ccc")
}

func TestLazyCtxLog(t *testing.T) {
	ctxBufLog := ctxBufLog{}
	l := log.NewLazyCtxLog(&ctxBufLog)
	l.Printf("aaa")
	require.Equal(t, "", ctxBufLog.buf)
	l.FlushCtx(context.Background())
	require.Contains(t, ctxBufLog.buf, "aaa")
}
