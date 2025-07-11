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

package throttle

// Noop defines a trivial throttle which does nothing.
type Noop struct{}

// NewNoop create a new Noop throttle.
func NewNoop() *Noop {
	return &Noop{}
}

// Allow Always return true
func (*Noop) Allow() bool { return true }

// OnSuccess empty implementation.
func (*Noop) OnSuccess() {}

// OnFailure empty implementation.
func (*Noop) OnFailure() {}
