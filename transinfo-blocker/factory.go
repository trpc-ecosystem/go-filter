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

// Package blocker 用于屏蔽调用下游的字段，避免登录态及其他敏感信息泄露问题
// @Title  factory.go
// @Description  解析trpc_go.yaml中plugin配置，初始化filter依赖的配置结构体
// @Author  radaren 2020.06.02
// @Update  radaren 2020.06.02
package blocker

import (
	"errors"
	"fmt"

	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginName    = "transinfo-blocker"
	pluginType    = "security"
	modeWhitelist = "whitelist"
	modeBlacklist = "blacklist"
	modeNone      = "none"
)

var cfg = &Config{}

// Config 插件配置
type Config struct {
	Default    *ListCfg            `yaml:"default"`      // Default 对所有client生效
	RPCNameCfg map[string]*ListCfg `yaml:"rpc_name_cfg"` // RPCNameCfg 对于特定的rpcname生效
}

// ListCfg 具体内容配置
type ListCfg struct {
	Mode string              `yaml:"mode"` // Mode "blocker, blacklist, none"
	Keys []string            `yaml:"keys"` // Keys blocker or blacklist key
	Set  map[string]struct{} `yaml:"-"`    // Set keys list to set
}

func init() {
	plugin.Register(pluginName, &TransinfoBlocker{})
}

// TransinfoBlocker 安全插件
type TransinfoBlocker struct{}

// Type transinfo-blocker 插件类型
func (t *TransinfoBlocker) Type() string {
	return pluginType
}

// Setup transinfo-blocker 实例初始化
func (t *TransinfoBlocker) Setup(name string, configDec plugin.Decoder) error {
	if configDec == nil {
		return errors.New("transinfo-blocker configDec nil")
	}

	cfg = &Config{}
	if err := configDec.Decode(cfg); err != nil {
		return err
	}
	if err := cfg.Default.parseBlockSet(); err != nil {
		return err
	}
	for _, v := range cfg.RPCNameCfg {
		if err := v.parseBlockSet(); err != nil {
			return err
		}
	}
	return nil
}

func (l *ListCfg) parseBlockSet() error {
	if l == nil {
		return nil
	}
	if l.Mode != modeNone &&
		l.Mode != modeWhitelist &&
		l.Mode != modeBlacklist {
		return fmt.Errorf("unknown mod: %s", l.Mode)
	}
	if l != nil {
		l.Set = make(map[string]struct{})
		for _, k := range l.Keys {
			l.Set[k] = struct{}{}
		}
	}
	return nil
}
