// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package mock

import (
	"encoding/base64"
	"time"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginName = "mock"
	pluginType = "tracing"
)

func init() {
	plugin.Register(pluginName, &Plugin{})
}

// Plugin mock trpc plugin implementation.
type Plugin struct{}

// Type mock trpc plugin type.
func (p *Plugin) Type() string {
	return pluginType
}

// Config mock plugin config.
type Config []*Item

// MockItem specific mock items.
// Deprecated: use Item instead.
type MockItem = Item

// Item Specific mock items
type Item struct {
	Method        string
	Retcode       int
	Retmsg        string
	Delay         int
	delay         time.Duration
	Timeout       bool
	Body          string
	data          []byte
	Serialization int // json jce pb
	Percent       int
}

// Setup mock instance initialization.
func (p *Plugin) Setup(name string, configDec plugin.Decoder) error {
	conf := Config{}
	if err := configDec.Decode(&conf); err != nil {
		return err
	}

	var opt []Option
	for _, mock := range conf {
		mock.delay = time.Millisecond * time.Duration(mock.Delay)
		if mock.Serialization != codec.SerializationTypeJSON {
			// When the serialization method is not json, use base64 to decode.
			decoded, err := base64.StdEncoding.DecodeString(mock.Body)
			if err != nil {
				return err
			}
			mock.data = decoded
		} else {
			mock.data = []byte(mock.Body)
		}
		opt = append(opt, WithMock(mock))
	}

	filter.Register(pluginName, nil, ClientFilter(opt...))
	return nil
}
