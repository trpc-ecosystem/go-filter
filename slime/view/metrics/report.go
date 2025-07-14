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

package metrics

import (
	"context"
	"strconv"
	"time"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"

	"trpc.group/trpc-go/trpc-filter/slime/view"
)

type counter interface {
	// Inc a counter should have a Inc method.
	Inc(name string, cnt int, tagPairs ...string)
}

type histogram interface {
	// Observe a histogram should have an Observe method like prometheus.
	Observe(name string, v float64, tagPairs ...string)
}

// Emitter defines a common metric interface.
type Emitter interface {
	counter
	histogram
}

const (
	// FQNAppRequest the number of requests triggered by user.
	FQNAppRequest = "appRequest"
	// FQNRealRequest the number of total requests sent to callee.
	FQNRealRequest = "realRequest"
	// FQNAppCostMs the cost of app requests.
	FQNAppCostMs = "appCostMs"
	// FQNRealCostMs the cost of real requests.
	FQNRealCostMs = "realCostMs"
)

const (
	// TagCaller is the caller name.
	TagCaller = "caller"
	// TagCallee is the callee name.
	TagCallee = "callee"
	// TagMethod is the callee method.
	TagMethod = "method"
	// TagAttempts gives how many attempts an app request has used.
	TagAttempts = "attempts"
	// TagErrCodes is the error code the request.
	TagErrCodes = "error_codes"
	// TagThrottled indicate whether is this request throttled.
	TagThrottled = "throttled"
	// TagInflight is inflight number of app request or indicate whether the real request is still inflight.
	TagInflight = "inflight"
	// TagNoMoreAttempt indicate whether the server issues no retry/hedging.
	TagNoMoreAttempt = "noMoreAttempt"
)

// TagNamesApp gives the order in which tagPairs are organized for FQNAppRequest or FQNAppCostMs.
var TagNamesApp = []string{
	TagCaller, TagCallee, TagMethod,
	TagAttempts, TagErrCodes, TagThrottled, TagInflight, TagNoMoreAttempt,
}

// TagNamesReal gives the order in which tagPairs are organized for FQNRealRequest or FQNRealCostMs.
var TagNamesReal = []string{
	TagCaller, TagCallee, TagMethod,
	TagErrCodes, TagInflight, TagNoMoreAttempt,
}

// Report report view.Stat.
type Report struct {
	counter   counter
	histogram histogram
}

// NewReport create a new Report from Emitter and tagPairs.
func NewReport(emitter Emitter, tagPairs ...string) *Report {
	return &Report{
		counter:   wrapCounter(emitter, tagPairs...),
		histogram: wrapHistogram(emitter, tagPairs...),
	}
}

// Report reports info of view.Stat. ctx is used to retrieve caller, callee and method.
func (r *Report) Report(ctx context.Context, stat view.Stat) {
	tags := r.tagsFromCtx(ctx)

	var noMoreAttempt bool
	for _, a := range stat.Attempts() {
		if a.NoMoreAttempt() {
			noMoreAttempt = true
		}

		// The order of tags must match TagNamesReal.
		realTags := append(tags,
			TagErrCodes, string(errs.Code(a.Error())),
			TagInflight, strconv.FormatBool(a.Inflight()),
			TagNoMoreAttempt, strconv.FormatBool(a.NoMoreAttempt()))

		r.counter.Inc(FQNRealRequest, 1, realTags...)

		var endTime = time.Now()
		if !a.Inflight() {
			endTime = a.End()
		}
		r.histogram.Observe(FQNRealCostMs, milliseconds(endTime.Sub(a.Start())), realTags...)
	}

	// The order of tags must match TagNamesApp.
	appTags := append(tags,
		TagAttempts, strconv.Itoa(len(stat.Attempts())),
		TagErrCodes, string(errs.Code(stat.Error())),
		TagThrottled, strconv.FormatBool(stat.Throttled()),
		TagInflight, strconv.Itoa(stat.InflightN()),
		TagNoMoreAttempt, strconv.FormatBool(noMoreAttempt))

	r.counter.Inc(FQNAppRequest, 1, appTags...)

	r.histogram.Observe(FQNAppCostMs, milliseconds(stat.Cost()), appTags...)
}

func (r *Report) tagsFromCtx(ctx context.Context) []string {
	var tags [6]string
	tags[0] = TagCaller
	tags[2] = TagCallee
	tags[4] = TagMethod

	msg := codec.Message(ctx)
	if caller := msg.CallerServiceName(); caller == "" {
		tags[1] = "unknown"
	} else {
		tags[1] = caller
	}
	if callee := msg.CalleeServiceName(); callee == "" {
		tags[3] = "unknown"
	} else {
		tags[3] = callee
	}
	if method := msg.CalleeMethod(); method == "" {
		tags[5] = "unknown"
	} else {
		tags[5] = method
	}

	return tags[:]
}

type counterWrapper struct {
	counter
	tagPairs []string
}

func wrapCounter(c counter, tagPairs ...string) *counterWrapper {
	return &counterWrapper{
		counter:  c,
		tagPairs: tagPairs,
	}
}

// Inc implement counter.Inc.
func (c *counterWrapper) Inc(name string, cnt int, tagPairs ...string) {
	c.counter.Inc(name, cnt, append(c.tagPairs, tagPairs...)...)
}

type histogramWrapper struct {
	histogram
	tagPairs []string
}

func wrapHistogram(h histogram, tagPairs ...string) *histogramWrapper {
	return &histogramWrapper{
		histogram: h,
		tagPairs:  tagPairs,
	}
}

// Observe implement counter.Observe.
func (h *histogramWrapper) Observe(name string, v float64, tagPairs ...string) {
	h.histogram.Observe(name, v, append(h.tagPairs, tagPairs...)...)
}

func milliseconds(t time.Duration) float64 {
	return float64(t/1e6) + float64(t%1e6)/1e6
}
