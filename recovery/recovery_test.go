// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package recovery

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServerFilter(t *testing.T) {
	Convey("TestServerFilter", t, func() {
		filter := ServerFilter(WithRecoveryHandler(defaultRecoveryHandler))
		succHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return &struct{}{}, nil
		}
		failHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("something wrong")
		}
		ctx := context.Background()
		Convey("test succ", func() {
			rsp, err := filter(ctx, nil, succHandler)
			So(err, ShouldBeNil)
			So(rsp, ShouldNotBeNil)
		})
		Convey("test fail", func() {
			rsp, err := filter(ctx, nil, failHandler)
			So(err, ShouldResemble, defaultRecoveryHandler(ctx, "something wrong"))
			So(rsp, ShouldBeNil)
		})
	})
}

func assertHandlerEquality(a RecoveryHandler, b RecoveryHandler) {
	ctx := context.Background()
	So(a(ctx, 12), ShouldResemble, b(ctx, 12))
	So(a(ctx, "Hello"), ShouldResemble, b(ctx, "Hello"))
	So(a(ctx, false), ShouldResemble, b(ctx, false))
}

func TestWithRecoveryHandler(t *testing.T) {
	Convey("TestWithRecoveryHandler", t, func() {
		option := WithRecoveryHandler(defaultRecoveryHandler)
		var opts options
		option(&opts)
		assertHandlerEquality(opts.rh, defaultRecoveryHandler)
	})
}
