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

// @Title  factory_test.go
// @Description  测试解析配置正确性，以及生成配置结构体正确性
// @Author  radaren 2020.06.02
// @Update  radaren 2020.06.02
package blocker

import (
	"testing"

	yaml "gopkg.in/yaml.v3"

	"trpc.group/trpc-go/trpc-go/plugin"
)

var testyaml = `
transinfo-blocker:
  default: # 默认客户端调用配置，所有rpc调用未配置rpc_name_cfg会使用这个
    mode: blacklist # none, whitelist, blacklist
    keys: 
      - oidb_header
  rpc_name_cfg: # 单独命令字调用客户端配置, 会对于这个命令字覆盖default
    /trpc.qq_news.user_info.UserInfo/HandleProcess:
      mode: whitelist
      keys: # mode=whitelist, keys为空则所有都不透传
       - trpc-trace_
    /trpc.qq_news.user_info.UserInfo/Call:
      mode: blacklist
      keys: 
        - trpc-trace`

// TestParseCfg 测试正常解析用例
func TestParseCfg(t *testing.T) {
	node := &struct {
		TransinfoBlocker yaml.Node `yaml:"transinfo-blocker"`
	}{}

	err := yaml.Unmarshal([]byte(testyaml), node)
	if err != nil {
		t.Error("unmarshal yaml failed:" + err.Error())
	}
	decoder := &plugin.YamlNodeDecoder{Node: &node.TransinfoBlocker}
	err = (&TransinfoBlocker{}).Setup("", decoder)
	if err != nil {
		t.Error("setup failed:" + err.Error())
	}

	err = (&TransinfoBlocker{}).Setup("", nil)
	if err == nil {
		t.Error("setup check failed:" + err.Error())
	}
}

var errYaml = `
transinfo-blocker:
  default: # 默认客户端调用配置，所有rpc调用未配置rpc_name_cfg会使用这个
    mode: none
    keys: key`

// TestErrorYaml 测试yaml解析错误用例
func TestErrorYaml(t *testing.T) {
	node := &struct {
		TransinfoBlocker yaml.Node `yaml:"transinfo-blocker"`
	}{}
	err := yaml.Unmarshal([]byte(errYaml), node)
	if err != nil {
		t.Error("unmarshal yaml failed:" + err.Error())
	}
	decoder := &plugin.YamlNodeDecoder{Node: &node.TransinfoBlocker}
	err = (&TransinfoBlocker{}).Setup("", decoder)
	if err == nil {
		t.Error("setup check failed:" + err.Error())
	}
}

var errDefaultYaml = `
transinfo-blocker:
  default: # 默认客户端调用配置，所有rpc调用未配置rpc_name_cfg会使用这个
    mode: errmod # none, whitelist, blacklist`

// TestErrorDefaultYaml 测试default解析错误用例
func TestErrorDefaultYaml(t *testing.T) {
	node := &struct {
		TransinfoBlocker yaml.Node `yaml:"transinfo-blocker"`
	}{}
	err := yaml.Unmarshal([]byte(errDefaultYaml), node)
	if err != nil {
		t.Error("unmarshal yaml failed:" + err.Error())
	}
	decoder := &plugin.YamlNodeDecoder{Node: &node.TransinfoBlocker}
	err = (&TransinfoBlocker{}).Setup("", decoder)
	if err == nil {
		t.Error("setup check failed:" + err.Error())
	}
}

var errRPCYaml = `
transinfo-blocker:
  rpc_name_cfg: # 单独命令字调用客户端配置, 会对于这个命令字覆盖default
    /trpc.qq_news.user_info.UserInfo/HandleProcess:
      mode: errmod
      keys: # mode=whitelist, keys为空则所有都不透传`

// TestErrorRPCYaml 测试rpcname错误用例
func TestErrorRPCYaml(t *testing.T) {
	node := &struct {
		TransinfoBlocker yaml.Node `yaml:"transinfo-blocker"`
	}{}
	err := yaml.Unmarshal([]byte(errRPCYaml), node)
	if err != nil {
		t.Error("unmarshal yaml failed:" + err.Error())
	}
	decoder := &plugin.YamlNodeDecoder{Node: &node.TransinfoBlocker}
	err = (&TransinfoBlocker{}).Setup("", decoder)
	if err == nil {
		t.Error("setup check failed:" + err.Error())
	}
}
