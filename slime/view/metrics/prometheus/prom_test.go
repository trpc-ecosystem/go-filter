//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package prom_test

import (
	"testing"

	"trpc.group/trpc-go/trpc-filter/slime/view/metrics"
	prom "trpc.group/trpc-go/trpc-filter/slime/view/metrics/prometheus"
)

// Prometheus cannot properly test in UT.
// We just try to call exported functions and make sure they are not panic.
// Note, the number of tag pairs must match metrics.TagNamesXxx.
func TestEmitter(t *testing.T) {
	tagsApp := []string{
		metrics.TagCaller, "caller_",
		metrics.TagCallee, "callee_",
		metrics.TagMethod, "method_",
		metrics.TagAttempts, "2",
		metrics.TagErrCodes, "0",
		metrics.TagThrottled, "false",
		metrics.TagInflight, "1",
		metrics.TagNoMoreAttempt, "false",
	}
	tagsReal := []string{
		metrics.TagCaller, "caller_",
		metrics.TagCallee, "callee_",
		metrics.TagMethod, "method_",
		metrics.TagErrCodes, "123",
		metrics.TagInflight, "false",
		metrics.TagNoMoreAttempt, "true",
	}
	m := prom.NewEmitter()
	m.Inc(metrics.FQNAppRequest, 1, tagsApp...)
	m.Inc(metrics.FQNRealRequest, 1, tagsReal...)
	m.Observe(metrics.FQNAppCostMs, 10, tagsApp...)
	m.Observe(metrics.FQNRealCostMs, 10, tagsReal...)
}
