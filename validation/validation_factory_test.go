// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const confOldFormat = `
server:
 filter:
  - validation
plugins:
  auth:
    validation:
      logfile:
        - true
`

const confNewFormat = `
server:
 filter:
  - validation
plugins:
  auth:
    validation:
      enable_error_log: true
      server_validate_err_code: 100101
      client_validate_err_code: 100102
`

func TestValidationPlugin_Type(t *testing.T) {
	p := &ValidationPlugin{}
	assert.Equal(t, pluginType, p.Type())
}

func readConf(conf string) plugin.Decoder {
	cfg := trpc.Config{}
	if err := yaml.Unmarshal([]byte(conf), &cfg); err != nil {
		return nil
	}
	validCfg := cfg.Plugins[pluginType][pluginName]
	return &validCfg
}

func TestValidationPlugin_Setup(t *testing.T) {
	type args struct {
		name      string
		configDec plugin.Decoder
	}
	tests := []struct {
		name    string
		p       *ValidationPlugin
		args    args
		wantErr bool
	}{
		{
			name:    "test succ with confOldFormat",
			p:       &ValidationPlugin{},
			args:    args{name: pluginName, configDec: readConf(confOldFormat)},
			wantErr: false,
		},
		{
			name:    "test succ with confNewFormat",
			p:       &ValidationPlugin{},
			args:    args{name: pluginName, configDec: readConf(confNewFormat)},
			wantErr: false,
		},
		{
			name:    "test succ with configDec nil",
			p:       &ValidationPlugin{},
			args:    args{name: pluginName, configDec: nil},
			wantErr: false,
		},
		{
			name:    "test err configDec decode error",
			p:       &ValidationPlugin{},
			args:    args{name: pluginName, configDec: &plugin.YamlNodeDecoder{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ValidationPlugin{}
			err := p.Setup(tt.args.name, tt.args.configDec)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
