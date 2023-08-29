// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package tvar

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/number"
	"trpc.group/trpc-go/trpc-filter/tvar/meterprovider"
)

func Test_handler(t *testing.T) {
	initMeterProvider()
	initRPCMetrics()

	snapshot = []meterprovider.ExportRecord{
		{
			AggregationKind: aggregation.SumKind,
			Sum:             number.Number(1),
		},
	}

	w := httptest.NewRecorder()
	handler(w, nil)

	actual := w.Body.String()
	expected := "[\"instrument server:, metric name:, aggregation:Sum, unit:, value:1\"]"
	assert.Equal(t, expected, actual)
}

func Test_ticker(t *testing.T) {
	assert.NotPanics(t, func() {
		c := make(chan time.Time, 1)
		c <- time.Now()

		tk := &time.Ticker{C: c}
		defer tk.Stop()

		go ticker(tk)
	})
}

func Test_updateServiceQPS(t *testing.T) {
	assert.NotPanics(t, func() {
		initMeterProvider()
		initRPCMetrics()

		updateServiceQPS()
	})
}

func Test_dumpRPCMetrics(t *testing.T) {
	initMeterProvider()
	initRPCMetrics()

	snapshot = nil
	assert.Equal(t, []string(nil), dumpRPCMetrics())

	snapshot = []meterprovider.ExportRecord{
		{
			AggregationKind: aggregation.HistogramKind,
			Histogram: aggregation.Buckets{
				Boundaries: []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				Counts:     []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
		{
			AggregationKind: aggregation.SumKind,
			Sum:             number.Number(1),
		},
	}
	assert.Equal(t, []string{
		"instrument server:, metric name:, " +
			"aggregation:Histogram, unit:, " +
			"value:[~0]:0 [~1]:1 [~2]:2 [~3]:3 [~4]:4 [~5]:5 [~6]:6 [~7]:7 [~8]:8 [~9]:9 ",
		"instrument server:, metric name:, aggregation:Sum, unit:, value:1",
	}, dumpRPCMetrics())

	tvarCfg.Percentile = defaultLatencyPercentile
	assert.Equal(t, []string{
		"instrument server:, metric name:, " +
			"aggregation:Histogram, unit:, " +
			"value:[~0]:0 [~1]:1 [~2]:2 [~3]:3 " +
			"[~4]:4 [~5]:5 [~6]:6 [~7]:7 [~8]:8 [~9]:9 p50:7.29 p99:15.60 p999:15.96 ",
		"instrument server:, metric name:, aggregation:Sum, unit:, value:1",
	}, dumpRPCMetrics())
}
