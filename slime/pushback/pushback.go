// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package pushback is a middleware that delays the response of a request.
package pushback

import (
	"time"

	"trpc.group/trpc-go/trpc-go/codec"
)

// MetaKey is key used in meta data of msg.
const MetaKey = "trpc-pushback-delay"

// FromMsg retrieve pushback delay from msg.
func FromMsg(msg codec.Msg) *time.Duration {
	if pushbackDelay, ok := msg.ClientMetaData()[MetaKey]; ok {
		if d, err := time.ParseDuration(string(pushbackDelay)); err == nil {
			return &d
		}
	}
	return nil
}
