// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package mock

import (
	"context"
	"testing"
	"time"

	"git.code.oa.com/trpc-go/trpc-go"
	"git.code.oa.com/trpc-go/trpc-go/codec"
	"git.code.oa.com/trpc-go/trpc-go/filter"
	"git.code.oa.com/trpc-go/trpc-go/plugin"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

const configInfo = `
plugins:
  tracing:
    mock:
      - method: /trpc.app.server.service/method1   # mock the specified interface, or mock all interfaces if none is specified
        delay: 10000  # Delay 10ms
        retcode: 111  # Simulation returns specific error codes
        retmsg: "error msg" # Simulation returns specific error messages
      - method: /trpc.app.server.service/method2   # mock the specified interface, or mock all interfaces if none is specified
        timeout: true # Simulate timeout failure
        percent: 10   # Triggered by 10% chance
      - method: /trpc.app.server.service/method3   # mock the specified interface, or mock all interfaces if none is specified
        body: '{"a":"aaa"}' #  body json data
        serialization: 2 # json serialization type
        percent: 10   # Triggered by 10% chance
`

func TestFilter_PluginType(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, pluginType, p.Type())
}

func TestPlugin_Setup(t *testing.T) {
	cfg := trpc.Config{}
	err := yaml.Unmarshal([]byte(configInfo), &cfg)
	assert.Nil(t, err)

	conf := cfg.Plugins[pluginType][pluginName]
	p := &Plugin{}
	err = p.Setup(pluginName, &plugin.YamlNodeDecoder{Node: &conf})
	assert.Nil(t, err)
}

func TestClientFilter(t *testing.T) {
	req := &struct{}{}
	rsp := &struct{}{}

	serialization := codec.SerializationTypeJSON
	tests := []struct {
		input     *Item
		assertion assert.ValueAssertionFunc
	}{
		{
			input:     &Item{Method: "method"},
			assertion: assert.Nil,
		},
		{
			input:     &Item{Timeout: true, Percent: 100},
			assertion: assert.NotNil,
		},
		{
			input:     &Item{Delay: 10, delay: 10 * time.Millisecond, Percent: 100},
			assertion: assert.NotNil,
		},
		{
			input:     &Item{Delay: 10, delay: time.Millisecond, Percent: 100},
			assertion: assert.Nil,
		},
		{
			input:     &Item{Retcode: 1, Percent: 100},
			assertion: assert.NotNil,
		},
		{
			input:     &Item{Body: "{", Serialization: serialization, data: []byte("{"), Percent: 100},
			assertion: assert.NotNil,
		},
		{
			input:     &Item{Body: "{}", Serialization: serialization, data: []byte("{}"), Percent: 100},
			assertion: assert.Nil,
		},
		{
			input:     &Item{Percent: 0},
			assertion: assert.Nil,
		},
	}
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer cancel()

		fc := filter.Chain{ClientFilter(WithMock(tt.input))}

		err := fc.Handle(ctx, &req, &rsp, noopHandler)
		tt.assertion(t, err, err)
	}
}

func noopHandler(ctx context.Context, req interface{}, rsp interface{}) error {
	return nil
}
