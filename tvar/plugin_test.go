// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package tvar

import (
	"errors"
	"testing"

	"git.code.oa.com/trpc-go/trpc-go"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const configInfo = `
plugins:
  apm:
    tvar:
      percentile:
        - p50
        - p90
        - p99
`

func TestPluginFactory_Type(t *testing.T) {
	assert.Equal(t, pluginType, (&tvarFactory{}).Type())
}

func Test_tvarFactory_Setup(t *testing.T) {
	cfg := trpc.Config{}
	err := yaml.Unmarshal([]byte(configInfo), &cfg)
	assert.Nil(t, err)

	tvarCfg := cfg.Plugins[pluginType][pluginName]
	tr := &tvarFactory{}

	err = tr.Setup(pluginName, &mockDecoder{err: errors.New("fake error")})
	assert.NotNil(t, err)

	err = tr.Setup(pluginName, &tvarCfg)
	assert.Nil(t, err)
}

type mockDecoder struct {
	err error
}

func (d *mockDecoder) Decode(interface{}) error {
	if d.err != nil {
		return d.err
	}
	return nil
}
