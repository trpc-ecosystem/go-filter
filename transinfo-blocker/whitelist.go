// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package blocker 用于屏蔽调用下游的字段，避免登录态及其他敏感信息泄露问题
// @Title  whitelist.go
// @Description  trpc_go.yaml里面client-filter具体配置，引入后初始化
// @Author  radaren 2020.06.02
// @Update  radaren 2020.06.02
package blocker

import (
	"context"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/filter"
)

func init() {
	filter.Register("transinfo-blocker", filter.NoopServerFilter, ClientFilter)
}

// ClientFilter 客户端调用屏蔽metadata内容
func ClientFilter(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
	ParseClientMetadata(ctx)
	return handler(ctx, req, rsp)
}

// ParseClientMetadata 修改metadata
func ParseClientMetadata(ctx context.Context) {
	msg := trpc.Message(ctx)
	metaData := msg.ClientMetaData()

	if len(metaData) == 0 { // 只有当前有传递给下游metaData的时候，才考虑这个逻辑
		return
	}

	transMetaData := make(map[string][]byte)
	callCfg := cfg.Default
	if rpcCfg, ok := cfg.RPCNameCfg[msg.ClientRPCName()]; ok {
		callCfg = rpcCfg
	}

	if callCfg == nil || callCfg.Mode == modeNone {
		return
	}

	for k, v := range metaData {
		if callCfg.Mode == modeWhitelist {
			if _, ok := callCfg.Set[k]; ok {
				transMetaData[k] = v
			}
		}

		if callCfg.Mode == modeBlacklist {
			if _, ok := callCfg.Set[k]; !ok {
				transMetaData[k] = v
			}
		}
	}
	msg.WithClientMetaData(transMetaData)
}
