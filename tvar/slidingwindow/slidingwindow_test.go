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
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlidingWindow(t *testing.T) {
	winsz := time.Millisecond * 100
	// new slidingwindow
	w := NewSlidingWindow(winsz)
	assert.NotNil(t, w.curr)
	assert.NotNil(t, w.prev)
	assert.Equal(t, winsz, w.size)

	// count
	assert.Zero(t, w.Count())

	// size
	assert.Equal(t, winsz, w.Size())

	// record
	w.Record()
	assert.Equal(t, int64(1), w.Count())

	// record
	time.Sleep(w.size)
	count := 5
	for i := 0; i < count; i++ {
		w.Record()
	}
	assert.Equal(t, int64(count), w.Count())
}

var sizes = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 500,
	time.Second,
}

// BenchmarkSlidingWindow_Record/window-100ms-8         	 9593324	       128.5 ns/op
// BenchmarkSlidingWindow_Record/window-500ms-8         	 9498463	       123.2 ns/op
// BenchmarkSlidingWindow_Record/window-1s-8            	 9815679	       118.4 ns/op
func BenchmarkSlidingWindow_Record(b *testing.B) {
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("window-%v", sz), func(b *testing.B) {
			w := NewSlidingWindow(sz)
			for i := 0; i < b.N; i++ {
				w.Record()
			}
		})
	}
}

// BenchmarkSlidingWindow_RecordCount/window-100ms-8         	 8293131	       131.6 ns/op
// BenchmarkSlidingWindow_RecordCount/window-500ms-8         	 9248320	       129.2 ns/op
// BenchmarkSlidingWindow_RecordCount/window-1s-8            	 9018114	       135.0 ns/op
func BenchmarkSlidingWindow_RecordCount(b *testing.B) {
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("window-%v", sz), func(b *testing.B) {
			w := NewSlidingWindow(sz)
			for i := 0; i < b.N; i++ {
				if i%10 != 0 {
					w.Record()
				} else {
					w.Count()
				}
			}
		})
	}
}

// BenchmarkSlidingWindow_RecordParrallel-8   	 7474735	       157.5 ns/op
func BenchmarkSlidingWindow_RecordParrallel(b *testing.B) {
	w := NewSlidingWindow(time.Second)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Record()
		}
	})
}

// BenchmarkSlidingWindow_RecordCountParrallel-8   	 5194528	       220.1 ns/op
func BenchmarkSlidingWindow_RecordCountParrallel(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%10)
	}

	var c int64
	w := NewSlidingWindow(time.Second)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddInt64(&c, 1)-1) % length
			v := inputs[i]
			if v != 0 {
				w.Record()
			} else {
				w.Count()
			}
		}
	})
}
