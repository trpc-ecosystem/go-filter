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

package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"

	"trpc.group/trpc-go/trpc-filter/slime/hedging"
	"trpc.group/trpc-go/trpc-filter/slime/view"
	. "trpc.group/trpc-go/trpc-filter/slime/view/metrics"
)

type emitter struct {
	Counters map[string][]struct {
		cnt      int
		tagPairs []string
	}
	Histograms map[string][]struct {
		v        float64
		tagPairs []string
	}
}

func newEmitter() *emitter {
	return &emitter{
		Counters: make(map[string][]struct {
			cnt      int
			tagPairs []string
		}),
		Histograms: make(map[string][]struct {
			v        float64
			tagPairs []string
		}),
	}
}

func (e *emitter) Inc(name string, cnt int, tagPairs ...string) {
	e.Counters[name] = append(e.Counters[name], struct {
		cnt      int
		tagPairs []string
	}{cnt: cnt, tagPairs: tagPairs})
}

func (e *emitter) Observe(name string, v float64, tagPairs ...string) {
	e.Histograms[name] = append(e.Histograms[name], struct {
		v        float64
		tagPairs []string
	}{v: v, tagPairs: tagPairs})
}

type stat struct {
	ValCost      time.Duration
	ValAttempts  []view.Attempt
	ValThrottled bool
	ValInflightN int
	ValError     error
}

func (s *stat) Cost() time.Duration {
	return s.ValCost
}

func (s *stat) Attempts() []view.Attempt {
	return s.ValAttempts
}

func (s *stat) Throttled() bool {
	return s.ValThrottled
}

func (s *stat) InflightN() int {
	return s.ValInflightN
}

func (s *stat) Error() error {
	return s.ValError
}

type attempt struct {
	ValStart, ValEnd time.Time
	ValError         error
	ValInflight      bool
	ValNoAttempt     bool
}

func (a *attempt) Start() time.Time {
	return a.ValStart
}

func (a *attempt) End() time.Time {
	return a.ValEnd
}

func (a *attempt) Error() error {
	return a.ValError
}

func (a *attempt) Inflight() bool {
	return a.ValInflight
}

func (a *attempt) NoMoreAttempt() bool {
	return a.ValInflight
}

func TestReport(t *testing.T) {
	caller := "caller_name"
	callee := "callee_name"
	method := "method_name"

	ctx, msg := codec.WithNewMessage(context.Background())
	msg.WithCallerServiceName(caller)
	msg.WithCalleeServiceName(callee)
	msg.WithCalleeMethod(method)

	now := time.Now()
	stat := stat{
		ValCost: time.Millisecond * 10,
		ValAttempts: []view.Attempt{
			&attempt{
				ValStart:     now,
				ValEnd:       now.Add(time.Millisecond * 15),
				ValError:     errs.New(errs.RetClientNetErr, ""),
				ValInflight:  false,
				ValNoAttempt: false,
			},
			&attempt{
				ValStart:     now.Add(time.Millisecond * 10),
				ValError:     hedging.InflightErr,
				ValInflight:  true,
				ValNoAttempt: false,
			},
			&attempt{
				ValStart:     now.Add(time.Millisecond * 20),
				ValEnd:       now.Add(time.Millisecond * 25),
				ValError:     errs.New(errs.RetServerOverload, "overloaded"),
				ValInflight:  false,
				ValNoAttempt: true,
			},
		},
		ValThrottled: true,
		ValInflightN: 1,
		ValError:     errs.New(errs.RetServerOverload, "overloaded"),
	}

	emitter := newEmitter()
	r := NewReport(emitter, "tagKey", "tagVal")
	r.Report(ctx, &stat)

	counters, ok := emitter.Counters[FQNAppRequest]
	require.True(t, ok)
	require.Len(t, counters, 1)
	require.Equal(t, 1, counters[0].cnt)
	require.Contains(t, counters[0].tagPairs, TagCallee)
	require.Contains(t, counters[0].tagPairs, TagCaller)

	histograms, ok := emitter.Histograms[FQNRealCostMs]
	require.True(t, ok)
	require.Len(t, histograms, 3)
}
