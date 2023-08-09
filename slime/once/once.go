// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package once defines the Once type which trivially wraps filter.ClientHandleFunc.
package once

import (
	"context"

	"trpc.group/trpc-go/trpc-go/filter"
)

// Once is a trivial wrapper of filter.ClientHandleFunc.
type Once struct{}

// New create a new Once.
func New() *Once {
	return &Once{}
}

// Invoke simply wraps handleFunc.
func (o *Once) Invoke(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) error {
	return f(ctx, req, rsp)
}
