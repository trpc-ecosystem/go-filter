// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package jwt

import (
	"fmt"
	"time"

	"git.code.oa.com/trpc-go/trpc-go/filter"
	"git.code.oa.com/trpc-go/trpc-go/plugin"
)

const (
	pluginName = "jwt"
	pluginType = "auth"
)

func init() {
	plugin.Register(pluginName, &pluginImp{})
}

// pluginImp jwt trpc 插件实现
type pluginImp struct{}

// Type jwt trpc插件类型
func (p *pluginImp) Type() string {
	return pluginType
}

// Config 插件配置
type Config struct {
	Secret       string   `yaml:"secret"`        // 签名使用的私钥
	Expired      int      `yaml:"expired"`       // 过期时间 seconds
	Issuer       string   `yaml:"issuer"`        // 发行人
	ExcludePaths []string `yaml:"exclude_paths"` // 跳过 jwt 鉴权的 paths, 如登陆接口
}

// Setup 插件实例初始化
func (p *pluginImp) Setup(name string, configDec plugin.Decoder) error {
	// 配置解析
	cfg := Config{}
	if err := configDec.Decode(&cfg); err != nil {
		return err
	}
	if cfg.Secret == "" {
		return fmt.Errorf("JWT secret not be empty")
	}
	// 未设置-默认为1小时过期
	var expired = time.Hour
	if cfg.Expired > 0 {
		expired = time.Duration(cfg.Expired) * time.Second
	}
	// slice 转换为 map
	var excludePathSet = make(map[string]bool)
	for _, s := range cfg.ExcludePaths {
		excludePathSet[s] = true
	}
	SetDefaultSigner(NewJwtSign([]byte(cfg.Secret), expired, cfg.Issuer))
	filter.Register(name, ServerFilter(WithExcludePathSet(excludePathSet)), nil)
	return nil
}
