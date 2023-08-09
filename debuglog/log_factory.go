// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package debuglog

import (
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginName = "debuglog"
	pluginType = "tracing"
)

// Register the plugin on init
func init() {
	plugin.Register(pluginName, &Plugin{})
}

// Plugin is the implement of the debuglog trpc plugin.
type Plugin struct {
}

// Type is the type of debuglog trpc plugin.
func (p *Plugin) Type() string {
	return pluginType
}

// Config is the congifuration for the debuglog trpc plugin.
type Config struct {
	LogType       string `yaml:"log_type"`
	ErrLogLevel   string `yaml:"err_log_level"`
	NilLogLevel   string `yaml:"nil_log_level"`
	ServerLogType string `yaml:"server_log_type"`
	ClientLogType string `yaml:"client_log_type"`
	EnableColor   *bool  `yaml:"enable_color"`
	Include       []*RuleItem
	Exclude       []*RuleItem
}

// get log func by log type
func getLogFunc(t string) LogFunc {
	switch t {
	case "simple":
		return SimpleLogFunc
	case "prettyjson":
		return PrettyJSONLogFunc
	case "json":
		return JSONLogFunc
	default:
		return DefaultLogFunc
	}
}

// Setup initializes the debuglog instance.
func (p *Plugin) Setup(name string, configDec plugin.Decoder) error {
	var conf Config
	err := configDec.Decode(&conf)
	if err != nil {
		return err
	}

	var serverOpt []Option
	var clientOpt []Option

	serverLogType := conf.LogType
	if conf.ServerLogType != "" {
		serverLogType = conf.ServerLogType
	}
	serverOpt = append(serverOpt, WithLogFunc(getLogFunc(serverLogType)))

	clientLogType := conf.LogType
	if conf.ClientLogType != "" {
		clientLogType = conf.ClientLogType
	}
	clientOpt = append(clientOpt, WithLogFunc(getLogFunc(clientLogType)))

	for _, in := range conf.Include {
		serverOpt = append(serverOpt, WithInclude(in))
		clientOpt = append(clientOpt, WithInclude(in))
	}
	for _, ex := range conf.Exclude {
		serverOpt = append(serverOpt, WithExclude(ex))
		clientOpt = append(clientOpt, WithExclude(ex))
	}

	clientOpt = append(clientOpt,
		WithNilLogLevelFunc(getLogLevelFunc(conf.NilLogLevel, "debug")),
		WithErrLogLevelFunc(getLogLevelFunc(conf.ErrLogLevel, "error")),
	)
	serverOpt = append(serverOpt,
		WithNilLogLevelFunc(getLogLevelFunc(conf.NilLogLevel, "debug")),
		WithErrLogLevelFunc(getLogLevelFunc(conf.ErrLogLevel, "error")),
	)
	if conf.EnableColor != nil {
		serverOpt = append(serverOpt, WithEnableColor(*conf.EnableColor))
		clientOpt = append(clientOpt, WithEnableColor(*conf.EnableColor))
	}

	// register server and client filter
	filter.Register(pluginName, ServerFilter(serverOpt...), ClientFilter(clientOpt...))

	return nil
}
