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

// Package slidingwindow provides an implementation of Sliding Window Algorithm.
//
// Design
// Let's take an limiter built on slidingwindow as an example to illustrate how the slidingwindow works.
//
// Suppose we have a limiter that permits 100 events per minute, and now the time comes at the "75s" point, then the
// internal sliding window will be as below:
//
// ```
//
//	Sliding Window
//
// |-------------------------------------|
// |  Previous Window | Current Window   |      window size: 60s
// |------------------|------------------|
// |        86        |        12        |
// |------------------|------------------|
// ^    ^             ^    ^             ^
// |    |             |    |             |
// 0s   15s           60s  75s           120s
//
// ```
//
// In this situation, the limiter has permitted 12 events during the current window, which started 15 seconds ago,
// and 86 events during the entire previous window. Then the count approximation during the sliding window can be
// calculated like this:
//
// ```
// count = 86 * ((60-15)/60) + 12
//
//	= 86 * 0.75 + 12
//	= 76.5 events
//
// ```
//
// In the example, count of requests is recorded in the slidingwindow. Actually we can see the request as an event,
// then we can record any events if we want.
package slidingwindow
