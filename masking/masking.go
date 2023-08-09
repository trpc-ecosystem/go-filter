// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package masking 敏感信息脱敏拦截器
package masking

import (
	"context"

	"git.code.oa.com/trpc-go/trpc-go/filter"
)

func init() {
	filter.Register(pluginName, ServerFilter(), nil)
}

// Masking 脱敏接口
type Masking interface {
	Masking()
}

// ServerFilter 服务端RPC调用自动校验req输入参数
func ServerFilter() filter.ServerFilter {
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (rsp interface{}, err error) {
		rsp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}
		DeepCheck(rsp)
		return rsp, nil
	}
}
