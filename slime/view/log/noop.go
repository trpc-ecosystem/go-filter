// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package log

import "context"

// NoopLog empty implementation.
type NoopLog struct{}

// Printf empty implementation.
func (NoopLog) Printf(string, ...interface{}) {}

// Flush empty implementation.
func (NoopLog) Flush() {}

// FlushCtx empty implementation.
func (NoopLog) FlushCtx(context.Context) {}
