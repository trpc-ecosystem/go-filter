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

package tvar

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"trpc.group/trpc-go/trpc-go"

	"trpc.group/trpc-go/trpc-filter/tvar/meterprovider"
	"trpc.group/trpc-go/trpc-filter/tvar/slidingwindow"
)

// rpc version
//
// TODO proposal未详细说明rpc_version统计项指的是什么版本，以及该版本应该统一如何获取
var rpcVersion = "unknown"

var (
	// default RPC latency boundaries in unit time.Millisecond
	//
	// TODO 配置文件中可配置
	defaultLatencyBoundaries = []float64{10, 20, 30, 50, 100, 200, 300, 500, 1000, 1500, 2000, 3000, 5000}

	provider metric.MeterProvider
	exporter *meterprovider.Exporter
	meter    metric.Meter

	// service side stats
	serviceConnectionNum asyncint64.Gauge
	serviceReqNum        syncint64.Counter
	serviceReqActiveNum  syncint64.UpDownCounter
	serviceRspNum        syncint64.Counter
	serviceReqSize       syncint64.Histogram // TODO filter中不便获取包尺寸
	serviceRspSize       syncint64.Histogram // TODO filter中不便获取包尺寸
	serviceErrNum        syncint64.Counter
	serviceBusiErrNum    syncint64.Counter

	// TODO 支持用户自定义的p1、p2、p3配置，当前支持了p50,p99,p999
	// TODO 支持统计avg、max、min，当前histogram不支持记录这些，如果支持了的话百分位数计算插值会更精准，但是似乎没什么必要
	serviceLatency            syncint64.Histogram
	serviceLatencyPercentiles = []int{50, 90, 99, 999, 9999} // user defined percentiles p1/p2/p3, or p999 p9999 by default

	// 使用asyncint64代替asyncfloat64，observe有个float64/uint64转换异常的问题
	serviceQPS  asyncint64.Gauge
	serviceQWin *slidingwindow.SlidingWindow

	// client side stats
	clientConnNum      asyncint64.Gauge
	clientReqNum       syncint64.Counter
	clientReqActiveNum syncint64.UpDownCounter
	clientRspNum       syncint64.Counter
	clientErrNum       syncint64.Counter

	// TODO 同serviceLatency
	clientLatency            syncint64.Histogram
	clientLatencyPercentiles = []int{50, 90, 99, 999, 9999} // user defined percentiles p1/p2/p3, or p999 p9999 by default
)

func initMeterProvider() {
	provider, exporter = meterprovider.NewOtelMeterProvider(
		meterprovider.WithExplicitBoundaries(defaultLatencyBoundaries),
		meterprovider.WithTemporalitySelector(aggregation.CumulativeTemporalitySelector()),
		// meterprovider.WithTemporalitySelector(aggregation.DeltaTemporalitySelector()),
	)

	cfg := trpc.GlobalConfig()
	name := fmt.Sprintf("trpc.%s.%s", cfg.Server.App, cfg.Server.Server)
	meter = provider.Meter(name)
}

func initRPCMetrics() {
	// init service side stats
	serviceConnectionNum, _ = meter.AsyncInt64().Gauge("rpc_service_connection_count",
		instrument.WithDescription("num of connection accepted and inused"),
		instrument.WithUnit(unit.Dimensionless))
	serviceReqNum, _ = meter.SyncInt64().Counter("rpc_service_req_total",
		instrument.WithDescription("num of req received"),
		instrument.WithUnit(unit.Dimensionless))
	serviceReqActiveNum, _ = meter.SyncInt64().UpDownCounter("rpc_service_req_active",
		instrument.WithDescription("num of req being processed"),
		instrument.WithUnit(unit.Dimensionless))
	serviceRspNum, _ = meter.SyncInt64().Counter("rpc_service_rsp_total",
		instrument.WithDescription("num of rsp sent"),
		instrument.WithUnit(unit.Dimensionless))
	serviceReqSize, _ = meter.SyncInt64().Histogram("rpc_service_req_avg_len",
		instrument.WithDescription("samples of req size"),
		instrument.WithUnit(unit.Bytes),
	)
	serviceRspSize, _ = meter.SyncInt64().Histogram("rpc_service_rsp_avg_len",
		instrument.WithDescription("samples of rsp size"),
		instrument.WithUnit(unit.Bytes))
	serviceErrNum, _ = meter.SyncInt64().Counter("rpc_service_error_total",
		instrument.WithDescription("num of errors"),
		instrument.WithUnit(unit.Dimensionless))
	serviceBusiErrNum, _ = meter.SyncInt64().Counter("rpc_service_business_error_total",
		instrument.WithDescription("num of business errors"),
		instrument.WithUnit(unit.Dimensionless))
	serviceLatency, _ = meter.SyncInt64().Histogram("rpc_service_latency",
		instrument.WithDescription("latency of serivce"),
		instrument.WithUnit(unit.Milliseconds))
	serviceQPS, _ = meter.AsyncInt64().Gauge("rpc_service_qps",
		instrument.WithDescription("qps of service"),
		instrument.WithUnit(unit.Dimensionless))
	serviceQWin = slidingwindow.NewSlidingWindow(time.Minute)

	// init client side stats

	clientConnNum, _ = meter.AsyncInt64().Gauge("rpc_client_connection_count",
		instrument.WithDescription("num of connection dialed and inused"),
		instrument.WithUnit(unit.Dimensionless))
	clientReqNum, _ = meter.SyncInt64().Counter("rpc_client_req_total",
		instrument.WithDescription("num of req sent"),
		instrument.WithUnit(unit.Dimensionless))
	clientReqActiveNum, _ = meter.SyncInt64().UpDownCounter("rpc_client_req_active",
		instrument.WithDescription("num of req waiting rsp"),
		instrument.WithUnit(unit.Dimensionless))
	clientRspNum, _ = meter.SyncInt64().Counter("rpc_client_rsp_total",
		instrument.WithDescription("num of rsp received"),
		instrument.WithUnit(unit.Dimensionless))
	clientErrNum, _ = meter.SyncInt64().Counter("rpc_client_error_total",
		instrument.WithDescription("num of errors"),
		instrument.WithUnit(unit.Dimensionless))
	clientLatency, _ = meter.SyncInt64().Histogram("rpc_client_latency",
		instrument.WithDescription("latency of client"),
		instrument.WithUnit(unit.Milliseconds))
}
