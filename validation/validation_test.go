// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package validation

import (
	"context"
	"errors"
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/http"
)

var (
	serverHandler = func(rsp interface{}, err error) filter.ServerHandleFunc {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			return rsp, err
		}
	}
	clientHandler = func(err error) filter.ClientHandleFunc {
		return func(ctx context.Context, req, rsp interface{}) error {
			return err
		}
	}
)

type errValidator struct{}

func (e *errValidator) Validate() error {
	return fmt.Errorf("err")
}

type succValidator struct{}

func (e *succValidator) Validate() error {
	return nil
}

func TestUnitServerFilter(t *testing.T) {
	Convey("TestUnitServerFilter with default config", t, func() {
		filter := ServerFilterWithOptions(defaultOptions)

		Convey("test invalid", func() {
			req := new(struct{})
			rsp, err := filter(context.Background(), req, serverHandler(&struct{}{}, nil))
			So(err, ShouldBeNil)
			So(rsp, ShouldNotBeNil)
		})
		Convey("test validate success", func() {
			req := new(succValidator)
			rsp, err := filter(context.Background(), req, serverHandler(&struct{}{}, nil))
			So(err, ShouldBeNil)
			So(rsp, ShouldNotBeNil)
		})
		Convey("test validate failed", func() {
			req := new(errValidator)
			rsp, err := filter(context.Background(), req, serverHandler(&struct{}{}, nil))
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
			So(errs.Code(err), ShouldEqual, errs.RetServerValidateFail)
		})
		Convey("test validate failed with ctx", func() {
			req := new(errValidator)
			r := &stdhttp.Request{
				URL: &url.URL{
					Path:     "path",
					RawQuery: "value",
				},
			}
			w := httptest.NewRecorder()
			w.Header().Set("Content-Type", "application/json")
			header := &http.Header{Request: r, Response: w}
			ctx := http.WithHeader(context.Background(), header)
			rsp, err := filter(ctx, req, serverHandler(nil, errors.New("err fake")))
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
			So(errs.Code(err), ShouldEqual, errs.RetServerValidateFail)
		})
	})

	Convey("TestUnitServerFilter with custom config", t, func() {
		filter := ServerFilter(
			WithErrorLog(true),
			WithServerValidateErrCode(-1),
		)

		Convey("test invalid", func() {
			req := new(struct{})
			rsp, err := filter(context.Background(), req, serverHandler(&struct{}{}, nil))
			So(err, ShouldBeNil)
			So(rsp, ShouldNotBeNil)
		})
		Convey("test validate success", func() {
			req := new(succValidator)
			rsp, err := filter(context.Background(), req, serverHandler(&struct{}{}, nil))
			So(err, ShouldBeNil)
			So(rsp, ShouldNotBeNil)
		})
		Convey("test validate failed", func() {
			req := new(errValidator)
			rsp, err := filter(context.Background(), req, serverHandler(&struct{}{}, nil))
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
			So(errs.Code(err), ShouldEqual, -1)
		})
		Convey("test validate failed with ctx", func() {
			req := new(errValidator)
			r := &stdhttp.Request{
				URL: &url.URL{
					Path:     "path",
					RawQuery: "value",
				},
			}
			w := httptest.NewRecorder()
			w.Header().Set("Content-Type", "application/json")
			header := &http.Header{Request: r, Response: w}
			ctx := http.WithHeader(context.Background(), header)
			rsp, err := filter(ctx, req, serverHandler(nil, errors.New("err fake")))
			So(err, ShouldNotBeNil)
			So(rsp, ShouldBeNil)
			So(errs.Code(err), ShouldEqual, -1)
		})
	})
}

func TestUnitClientFilter(t *testing.T) {
	Convey("TestUnitClientFilter with default config", t, func() {
		filter := ClientFilterWithOptions(defaultOptions)

		Convey("test handler err", func() {
			rsp := new(succValidator)
			err := filter(context.Background(), nil, rsp, clientHandler(errors.New("err fake")))
			So(err, ShouldNotBeNil)
			So(errs.Code(err), ShouldEqual, errs.RetUnknown)
		})
		Convey("test invalid", func() {
			rsp := new(struct{})
			err := filter(context.Background(), nil, rsp, clientHandler(nil))
			So(err, ShouldBeNil)
		})
		Convey("test validate failed", func() {
			rsp := new(errValidator)
			err := filter(context.Background(), nil, rsp, clientHandler(nil))
			So(err, ShouldNotBeNil)
			So(errs.Code(err), ShouldEqual, errs.RetClientValidateFail)
		})
		Convey("test validate success", func() {
			rsp := new(succValidator)
			err := filter(context.Background(), nil, rsp, clientHandler(nil))
			So(err, ShouldBeNil)
		})
	})

	Convey("TestUnitClientFilter with custom config", t, func() {
		filter := ClientFilter(
			WithLogfile(true),
			WithClientValidateErrCode(-1),
		)

		Convey("test handler err", func() {
			rsp := new(succValidator)
			err := filter(context.Background(), nil, rsp, clientHandler(errors.New("err fake")))
			So(err, ShouldNotBeNil)
			So(errs.Code(err), ShouldEqual, errs.RetUnknown)
		})
		Convey("test invalid", func() {
			rsp := new(struct{})
			err := filter(context.Background(), nil, rsp, clientHandler(nil))
			So(err, ShouldBeNil)
		})
		Convey("test validate failed", func() {
			rsp := new(errValidator)
			err := filter(context.Background(), nil, rsp, clientHandler(nil))
			So(err, ShouldNotBeNil)
			So(errs.Code(err), ShouldEqual, -1)
		})
		Convey("test validate success", func() {
			rsp := new(succValidator)
			err := filter(context.Background(), nil, rsp, clientHandler(nil))
			So(err, ShouldBeNil)
		})
	})
}
