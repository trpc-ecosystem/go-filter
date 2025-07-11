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

package meterprovider

import "go.opentelemetry.io/otel/sdk/metric/export/aggregation"

type config struct {
	temporalitySelector aggregation.TemporalitySelector
	boundaries          []float64
}

func newConfig(opts ...Option) config {
	cfg := config{
		temporalitySelector: aggregation.CumulativeTemporalitySelector(),
	}
	for _, opt := range opts {
		cfg = opt.apply(cfg)
	}
	return cfg
}

// Option allow for control of details of the TestMeterProvider created.
type Option interface {
	apply(config) config
}

type functionOption func(config) config

func (f functionOption) apply(cfg config) config {
	return f(cfg)
}

// WithTemporalitySelector allows for the use of either cumulative (default) or
// delta metrics.
//
// Warning: the current SDK does not convert async instruments into delta
// temporality.
func WithTemporalitySelector(ts aggregation.TemporalitySelector) Option {
	return functionOption(func(cfg config) config {
		if ts == nil {
			return cfg
		}
		cfg.temporalitySelector = ts
		return cfg
	})
}

// WithExplicitBoundaries allows for the explicit specified boundaries for histogram.
func WithExplicitBoundaries(boundaries []float64) Option {
	return functionOption(func(cfg config) config {
		cfg.boundaries = boundaries
		return cfg
	})
}
