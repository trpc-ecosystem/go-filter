// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package slidingwindow

import (
	"sync"
	"time"
)

// SlidingWindow sliding window consists two windows `curr` and `prev`,
// the window is advanced when recording events.
type SlidingWindow struct {
	size time.Duration

	mu sync.Mutex

	curr *window
	prev *window
}

// NewSlidingWindow creates a new slidingwindow
func NewSlidingWindow(size time.Duration) *SlidingWindow {
	currWin := newLocalWindow()

	// The previous window is static (i.e. no add changes will happen within it),
	// so we always create it as an instance of window.
	//
	// In this way, the whole limiter, despite containing two windows, now only
	// consumes at most one goroutine for the possible sync behaviour within
	// the current window.
	prevWin := newLocalWindow()

	return &SlidingWindow{
		size: size,
		curr: currWin,
		prev: prevWin,
	}
}

// Size returns the time duration of one window size. Note that the size
// is defined to be read-only, if you need to change the size,
// create a new limiter with a new size instead.
func (sw *SlidingWindow) Size() time.Duration {
	return sw.size
}

// Record report whether a event may happen at time now.
func (sw *SlidingWindow) Record() {
	sw.RecordN(time.Now(), 1)
}

// RecordN reports whether n events may happen at time now.
func (sw *SlidingWindow) RecordN(now time.Time, n int64) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.advance(now)
	sw.curr.AddCount(n)
}

// Count counts new window size.
func (sw *SlidingWindow) Count() int64 {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.advance(now)

	elapsed := now.Sub(sw.curr.Start())
	weight := float64(sw.size-elapsed) / float64(sw.size)
	count := int64(weight*float64(sw.prev.Count())) + sw.curr.Count()

	return count
}

// advance updates the current/previous windows resulting from the passage of time.
func (sw *SlidingWindow) advance(now time.Time) {
	// Calculate the start boundary of the expected current-window.
	newCurrStart := now.Truncate(sw.size)

	diffSize := newCurrStart.Sub(sw.curr.Start()) / sw.size

	// Fast path, the same window
	if diffSize == 0 {
		return
	}

	// Slow path, the current-window is at least one-window-size behind the expected one.

	// The new current-window always has zero count.
	sw.curr.Reset(newCurrStart, 0)

	// reset previous window
	newPrevCount := int64(0)
	if diffSize == 1 {
		// The new previous-window will overlap with the old current-window,
		// so it inherits the count.
		//
		// Note that the count here may be not accurate, since it is only a
		// SNAPSHOT of the current-window's count, which in itself tends to
		// be inaccurate due to the asynchronous nature of the sync behaviour.
		newPrevCount = sw.curr.Count()
	}
	sw.prev.Reset(newCurrStart.Add(-sw.size), newPrevCount)
}
