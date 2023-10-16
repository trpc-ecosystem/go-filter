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

package masking

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const confNoLogfile = `
server:
 filter:
  - validation
plugins:
  auth:
    validation:
`

func readConf(conf string) plugin.Decoder {
	cfg := trpc.Config{}
	err := yaml.Unmarshal([]byte(conf), &cfg)
	if err != nil {
		return nil
	}
	validCfg := cfg.Plugins[pluginType][pluginName]
	return &validCfg
}

func TestMaskingPlugin_Type(t *testing.T) {
	p := &MaskingPlugin{}
	assert.Equal(t, pluginType, p.Type())
}

func TestPlugin_Setup(t *testing.T) {
	type args struct {
		name      string
		configDec plugin.Decoder
	}
	tests := []struct {
		name    string
		p       *MaskingPlugin
		args    args
		wantErr bool
	}{
		{"test succ no logfile", &MaskingPlugin{}, args{name: pluginName, configDec: readConf(confNoLogfile)}, false},
		{"test err configDec nil", &MaskingPlugin{}, args{name: pluginName, configDec: nil}, true},
		{"test err configDec decode error", &MaskingPlugin{}, args{name: pluginName,
			configDec: &plugin.YamlNodeDecoder{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MaskingPlugin{}
			err := p.Setup(tt.args.name, tt.args.configDec)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
