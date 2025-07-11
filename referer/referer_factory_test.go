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

package referer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go"
)

const configInfo = `
server:
  filter:
    - referer
plugins:
  auth:
    referer:
      /trpc.app.server.service/method:
        - qq.com
`

// TestPlugin_Setup TestPlugin_Setup
func TestPlugin_Setup(t *testing.T) {
	cfg := trpc.Config{}
	err := yaml.Unmarshal([]byte(configInfo), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, len(cfg.Server.Filter), 1)

	corsCfg := cfg.Plugins[pluginType][pluginName]
	p := &Plugin{}
	err = p.Setup(pluginName, &corsCfg)
	assert.Nil(t, err)

	err = p.Setup(pluginName, nil)
	assert.NotNil(t, err)
}
