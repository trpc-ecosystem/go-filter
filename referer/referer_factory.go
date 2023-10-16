//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package referer

import (
	"errors"

	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginName = "referer"
	pluginType = "auth"
)

func init() {
	plugin.Register(pluginName, &Plugin{})
}

// Plugin  插件实现
type Plugin struct{}

// Type Plugin trpc插件类型
func (p *Plugin) Type() string {
	return pluginType
}

// Setup Referer实例初始化
func (p *Plugin) Setup(name string, configDec plugin.Decoder) error {
	// 配置解析
	if configDec == nil {
		return errors.New("referer writer decoder empty")
	}

	conf := make(map[string][]string)
	if err := configDec.Decode(&conf); err != nil {
		return err
	}

	var opt []Option
	for methodName, allowDomain := range conf {
		opt = append(opt, WithRefererDomain(methodName, allowDomain...))
	}

	filter.Register(pluginName, ServerFilter(opt...), nil)
	return nil
}
