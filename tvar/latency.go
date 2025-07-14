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

import "go.opentelemetry.io/otel/sdk/metric/export/aggregation"

type latencyHistogram struct {
	aggregation.Buckets
}

func (h *latencyHistogram) Percentile(p float64) float64 {
	totalCount := h.TotalCount()
	findValue := p * float64(totalCount)
	insertIdx := 0

	sum := float64(0)
	for i, v := range h.Counts {
		sum += float64(v)
		if sum >= findValue {
			insertIdx = i
			break
		}
	}
	// 缺少最小值，以上界估计
	if insertIdx == 0 {
		return h.Boundaries[0]
	}

	// 缺少最大值, 以下界估计
	if insertIdx == len(h.Boundaries) {
		return h.Boundaries[len(h.Boundaries)-1]
	}

	// 计算insertIdx之前的桶的占比
	countLtInsertIdx := uint64(0)
	countEqInsertIdx := uint64(0)

	for i := 0; i < insertIdx; i++ {
		countLtInsertIdx += h.Counts[i]
	}
	countEqInsertIdx = countLtInsertIdx + h.Counts[insertIdx]

	percentLowerBound := float64(countLtInsertIdx) / float64(totalCount)
	percentUpperBound := float64(countEqInsertIdx) / float64(totalCount)

	interpolation := (p - percentLowerBound) / (percentUpperBound - percentLowerBound)
	return (h.Boundaries[insertIdx-1]) * (1 + interpolation)
}

func (h *latencyHistogram) TotalCount() uint64 {
	var sum uint64
	for _, v := range h.Counts {
		sum += v
	}
	return sum
}
