// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package prom is a prometheus reporter.
package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"trpc.group/trpc-go/trpc-filter/slime/view/metrics"
)

const (
	subSystem = "slime"
)

// Emitter is a prometheus Reporter.
type Emitter struct {
	appReq   *prometheus.CounterVec
	realReq  *prometheus.CounterVec
	appCost  *prometheus.HistogramVec
	realCost *prometheus.HistogramVec
}

// NewEmitter create a new emitter.
func NewEmitter() *Emitter {
	return &Emitter{
		appReq: promauto.NewCounterVec(
			prometheus.CounterOpts{Subsystem: subSystem, Name: metrics.FQNAppRequest},
			metrics.TagNamesApp),
		realReq: promauto.NewCounterVec(
			prometheus.CounterOpts{Subsystem: subSystem, Name: metrics.FQNRealRequest},
			metrics.TagNamesReal),
		appCost: promauto.NewHistogramVec(
			prometheus.HistogramOpts{Subsystem: subSystem, Name: metrics.FQNAppCostMs},
			metrics.TagNamesApp),
		realCost: promauto.NewHistogramVec(
			prometheus.HistogramOpts{Subsystem: subSystem, Name: metrics.FQNRealCostMs},
			metrics.TagNamesReal),
	}
}

// Inc increases name by cnt with tagPairs.
func (e *Emitter) Inc(name string, cnt int, tagPairs ...string) {
	switch name {
	case metrics.FQNAppRequest:
		e.appReq.With(tagPairs2Labels(tagPairs)).Add(float64(cnt))
	case metrics.FQNRealRequest:
		e.realReq.With(tagPairs2Labels(tagPairs)).Add(float64(cnt))
	default:
	}
}

// Observe increases name by v with tagPairs.
func (e *Emitter) Observe(name string, v float64, tagPairs ...string) {
	switch name {
	case metrics.FQNAppCostMs:
		e.appCost.With(tagPairs2Labels(tagPairs)).Observe(v)
	case metrics.FQNRealCostMs:
		e.realCost.With(tagPairs2Labels(tagPairs)).Observe(v)
	default:
	}
}

func tagPairs2Labels(tagPairs []string) prometheus.Labels {
	labels := make(prometheus.Labels)
	for i := 0; i+1 < len(tagPairs); i += 2 {
		labels[tagPairs[i]] = tagPairs[i+1]
	}
	return labels
}
