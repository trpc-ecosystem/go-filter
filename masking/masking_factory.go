// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package masking 敏感信息脱敏拦截器
package masking

import (
	"errors"

	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginName = "masking"
	pluginType = "auth"
)

func init() {
	plugin.Register(pluginName, &MaskingPlugin{})
}

// MaskingPlugin masking trpc 插件实现
type MaskingPlugin struct {
}

// Type masking trpc插件类型
func (p *MaskingPlugin) Type() string {
	return pluginType
}

// Setup masking实例初始化
func (p *MaskingPlugin) Setup(_ string, configDec plugin.Decoder) error {

	// 配置解析
	if configDec == nil {
		return errors.New("masking decoder empty")
	}
	conf := make(map[string][]string)
	err := configDec.Decode(&conf)
	if err != nil {
		return err
	}

	sf := ServerFilter()

	filter.Register(pluginName, sf, nil)
	return nil
}
