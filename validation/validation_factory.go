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

package validation

import (
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginName = "validation"
	pluginType = "auth"
)

func init() {
	plugin.Register(pluginName, &ValidationPlugin{})
}

// ValidationPlugin implements the trpc validation plugin.
type ValidationPlugin struct{} //nolint:revive

// Type validation trpc plugin type.
func (p *ValidationPlugin) Type() string {
	return pluginType
}

// Setup initializes the validation plugin instance.
func (p *ValidationPlugin) Setup(name string, configDec plugin.Decoder) error {
	o := defaultOptions

	// When the configuration is not empty,
	// the default values are overridden during configuration parsing.
	if configDec != nil {
		if err := configDec.Decode(&o); err != nil {
			return err
		}
	}

	// Compatible with the old version of the configuration format and make adjustments
	if !o.EnableErrorLog && len(o.LogFile) != 0 && o.LogFile[0] {
		o.EnableErrorLog = true
	}

	filter.Register(pluginName, ServerFilterWithOptions(o), ClientFilterWithOptions(o))
	return nil
}
