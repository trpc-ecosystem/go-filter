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

// Package view defines common interfaces about visualization of retry/hedging.
package view

import (
	"time"
)

// Stat defines a stat each retry/hedging must have.
type Stat interface {
	Cost() time.Duration
	Attempts() []Attempt
	Throttled() bool
	InflightN() int
	Error() error
}

// Attempt defines the stat of each retry/hedging attempt.
type Attempt interface {
	Start() time.Time
	End() time.Time
	Error() error
	Inflight() bool
	NoMoreAttempt() bool
}
