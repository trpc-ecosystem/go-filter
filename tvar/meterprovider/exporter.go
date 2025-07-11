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

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/number"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

// Exporter is a manually collected exporter for testing the SDK.  It does not
// satisfy the `export.Exporter` interface because it is not intended to be
// used with the periodic collection of the SDK, instead the test should
// manually call `Collect()`
//
// Exporters are not thread safe, and should only be used for testing.
type Exporter struct {
	// records contains the last metrics collected.
	records []ExportRecord
	mux     *sync.RWMutex

	controller          *controller.Controller
	temporalitySelector aggregation.TemporalitySelector
}

// NewOtelMeterProvider creates a MeterProvider and Exporter to be used in tests.
func NewOtelMeterProvider(opts ...Option) (metric.MeterProvider, *Exporter) {
	cfg := newConfig(opts...)

	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(histogram.WithExplicitBoundaries(cfg.boundaries)),
			cfg.temporalitySelector,
		),
		controller.WithCollectPeriod(0),
	)
	exp := &Exporter{
		mux:                 new(sync.RWMutex),
		controller:          c,
		temporalitySelector: cfg.temporalitySelector,
	}

	return c, exp
}

// Library is the same as "sdk/instrumentation".Library but there is
// a package cycle to use it so it is redeclared here.
type Library struct {
	InstrumentationName    string
	InstrumentationVersion string
	SchemaURL              string
}

// ExportRecord represents one collected datapoint from the Exporter.
type ExportRecord struct {
	InstrumentName         string
	InstrumentationLibrary Library
	Attributes             []attribute.KeyValue
	AggregationKind        aggregation.Kind
	NumberKind             number.Kind
	Sum                    number.Number
	Count                  uint64
	Histogram              aggregation.Buckets
	LastValue              number.Number
	Unit                   unit.Unit
}

// var descTpl = `instrument library:%s, version:%s, schemaURL:%s instrument name:%s, aggregation:%s, value:%v`
var descTpl = `instrument server:%s, metric name:%s, aggregation:%s, unit:%s, value:%v`

// String implements Stringer to custom print.
func (r ExportRecord) String() string {
	args := []interface{}{
		r.InstrumentationLibrary.InstrumentationName,
		// TODO 如果构建时指定了`go build -vcs'，可以尝试提取版本信息
		// r.InstrumentationLibrary.InstrumentationVersion,
		// r.InstrumentationLibrary.SchemaURL,
		r.InstrumentName,
		r.AggregationKind,
		r.Unit,
	}

	switch r.AggregationKind {
	case aggregation.SumKind:
		args = append(args, r.Sum)
	case aggregation.HistogramKind:
		sb := &strings.Builder{}
		for i, b := range r.Histogram.Boundaries {
			fmt.Fprintf(sb, "[~%v]:%v ", b, r.Histogram.Counts[i])
		}
		args = append(args, sb.String())
	case aggregation.LastValueKind:
		args = append(args, r.LastValue)
	default:
		return "not supported"
	}

	return fmt.Sprintf(descTpl, args...)
}

// Collect triggers the SDK's collect methods and then aggregates the data into
// ExportRecords.  This will overwrite any previous collected metrics.
func (e *Exporter) Collect(ctx context.Context) error {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.records = []ExportRecord{}

	err := e.controller.Collect(ctx)
	if err != nil {
		return err
	}

	return e.controller.ForEach(func(l instrumentation.Library, r export.Reader) error {
		lib := Library{
			InstrumentationName:    l.Name,
			InstrumentationVersion: l.Version,
			SchemaURL:              l.SchemaURL,
		}

		return r.ForEach(e.temporalitySelector, func(rec export.Record) error {
			record := ExportRecord{
				InstrumentName:         rec.Descriptor().Name(),
				InstrumentationLibrary: lib,
				Attributes:             rec.Attributes().ToSlice(),
				AggregationKind:        rec.Aggregation().Kind(),
				NumberKind:             rec.Descriptor().NumberKind(),
				Unit:                   rec.Descriptor().Unit(),
			}

			var err error
			switch agg := rec.Aggregation().(type) {
			case aggregation.Histogram:
				record.AggregationKind = aggregation.HistogramKind
				record.Histogram, err = agg.Histogram()
				if err != nil {
					return err
				}
				record.Sum, err = agg.Sum()
				if err != nil {
					return err
				}
				record.Count, err = agg.Count()
				if err != nil {
					return err
				}
			case aggregation.Count:
				record.Count, err = agg.Count()
				if err != nil {
					return err
				}
			case aggregation.LastValue:
				record.LastValue, _, err = agg.LastValue()
				if err != nil {
					return err
				}
			case aggregation.Sum:
				record.Sum, err = agg.Sum()
				if err != nil {
					return err
				}
			}

			e.records = append(e.records, record)
			return nil
		})
	})
}

// GetRecords returns all records found by the SDK.
func (e *Exporter) GetRecords() []ExportRecord {
	e.mux.RLock()
	defer e.mux.RUnlock()

	records := make([]ExportRecord, len(e.records))
	copy(records, e.records)
	return records
}

// ErrNotFound 记录未找到
var ErrNotFound = fmt.Errorf("record not found")

// GetByName returns the first Record with a matching instrument name.
func (e *Exporter) GetByName(name string) (ExportRecord, error) {
	e.mux.RLock()
	defer e.mux.RUnlock()

	for _, rec := range e.records {
		if rec.InstrumentName == name {
			return rec, nil
		}
	}
	return ExportRecord{}, ErrNotFound
}

// GetByNameAndAttributes returns the first Record with a matching name and the sub-set of attributes.
func (e *Exporter) GetByNameAndAttributes(name string, attributes []attribute.KeyValue) (ExportRecord, error) {
	e.mux.RLock()
	defer e.mux.RUnlock()

	for _, rec := range e.records {
		if rec.InstrumentName == name && subSet(attributes, rec.Attributes) {
			return rec, nil
		}
	}
	return ExportRecord{}, ErrNotFound
}

// subSet returns true if attributesA is a subset of attributesB.
func subSet(attributesA, attributesB []attribute.KeyValue) bool {
	b := attribute.NewSet(attributesB...)

	for _, kv := range attributesA {
		if v, found := b.Value(kv.Key); !found || v != kv.Value {
			return false
		}
	}
	return true
}
