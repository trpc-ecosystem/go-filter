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

package slidingwindow

import (
	"time"
)

// window represents a window that ignores sync behavior entirely
// and only stores counters in memory.
type window struct {
	// The start boundary (timestamp in nanoseconds) of the window.
	// [start, start + size)
	start int64

	// The total count of events happened in the window.
	count int64
}

func newLocalWindow() *window {
	return &window{}
}

func (w *window) Start() time.Time {
	return time.Unix(0, w.start)
}

func (w *window) Count() int64 {
	return w.count
}

func (w *window) AddCount(n int64) {
	w.count += n
}

func (w *window) Reset(s time.Time, c int64) {
	w.start = s.UnixNano()
	w.count = c
}
