// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package cpmsg_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"trpc.group/trpc-go/trpc-go/codec"

	"trpc.group/trpc-go/trpc-filter/slime/cpmsg"
)

type head struct {
	X int
	Y string
	Z map[int]string
}

func TestCopyMsg(t *testing.T) {
	var src, dst codec.Msg
	require.NotNil(t, cpmsg.CopyMsg(dst, src), "src and dst should be non nil")

	src = codec.Message(context.Background())
	dst = codec.Message(context.Background())
	src.WithCalleeApp("app")
	require.Nil(t, cpmsg.CopyMsg(dst, src))
	require.Equal(t, src.CalleeApp(), dst.CalleeApp())

	src.WithClientReqHead(&head{X: 1, Y: "y", Z: map[int]string{1: "1"}})
	dst = codec.Message(context.Background())
	require.Nil(t, cpmsg.CopyMsg(dst, src))
	require.Equal(t, src.ClientReqHead(), dst.ClientReqHead())

	dst = codec.Message(context.Background())
	reqH := head{X: 2, Y: "yy", Z: map[int]string{2: "2"}}
	dst.WithClientReqHead(&reqH)
	require.Nil(t, cpmsg.CopyMsg(dst, src))
	require.Equal(t, src.ClientReqHead(), dst.ClientReqHead())
	require.Equal(t, src.ClientReqHead(), &reqH)

	src.WithClientRspHead(&head{X: -1, Y: "y", Z: map[int]string{-1: "-1"}})
	dst = codec.Message(context.Background())
	require.Nil(t, cpmsg.CopyMsg(dst, src))
	require.Equal(t, src.ClientRspHead(), dst.ClientRspHead())

	dst = codec.Message(context.Background())
	rspH := head{X: -2, Y: "yy", Z: map[int]string{-2: "-2"}}
	dst.WithClientRspHead(&rspH)
	require.Nil(t, cpmsg.CopyMsg(dst, src))
	require.Equal(t, src.ClientRspHead(), dst.ClientRspHead())
	require.Equal(t, src.ClientRspHead(), &rspH)
}
