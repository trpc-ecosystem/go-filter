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

package metrics

import (
	"context"

	"trpc.group/trpc-go/trpc-filter/slime/view"
)

// Noop empty implementation.
type Noop struct{}

// Report does nothing.
func (Noop) Report(context.Context, view.Stat) {}

// Inc implements Emitter and does nothing.
func (Noop) Inc(name string, cnt int, tagPairs ...string) {}

// Observe implements Emitter and does nothing.
func (Noop) Observe(name string, v float64, tagPairs ...string) {}
