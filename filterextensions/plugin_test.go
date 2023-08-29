// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package filterextensions_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-filter/filterextensions"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

func TestServiceMethodFilters_Type(t *testing.T) {
	f := plugin.Get(filterextensions.PluginType, filterextensions.PluginName)
	require.NotNil(t, f)
	require.Equal(t, filterextensions.PluginType, f.Type())
}

func TestServiceMethodFilters_Setup(t *testing.T) {
	f := plugin.Get(filterextensions.PluginType, filterextensions.PluginName)
	require.NotNil(t, f)

	var cmfCalled bool
	cmf := func(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
		cmfCalled = true
		return nil
	}
	var smfCalled bool
	smf := func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (interface{}, error) {
		smfCalled = true
		return nil, nil
	}

	dec := yaml.NewDecoder(bytes.NewReader([]byte(yamlCfg)))
	require.NotNil(t, f.Setup(filterextensions.PluginName, dec), "cmf or smf not registered")
	filter.Register("cmf", nil, cmf)
	filter.Register("smf", smf, nil)
	dec = yaml.NewDecoder(bytes.NewReader([]byte(yamlCfg)))
	require.Nil(t, f.Setup(filterextensions.PluginName, dec))

	serverFilters := filter.GetServer(filterextensions.MethodFilters)
	require.NotNil(t, serverFilters)
	clientFilters := filter.GetClient(filterextensions.MethodFilters)
	require.NotNil(t, clientFilters)

	noopClientHandler := func(ctx context.Context, req, rsp interface{}) error { return nil }
	testClientFilters := func(filters filter.ClientFilter, called *bool) {
		ctx := trpc.BackgroundContext()
		msg := trpc.Message(ctx)
		msg.WithCalleeServiceName("unknown-server-service")
		require.Nil(t, filters(ctx, nil, nil, noopClientHandler))
		require.False(t, *called, "unknown service")
		msg.WithCalleeServiceName("s_a")
		require.Nil(t, filters(ctx, nil, nil, noopClientHandler))
		require.False(t, *called, "missing method")
		msg.WithCalleeMethod("unknown-method")
		require.Nil(t, filters(ctx, nil, nil, noopClientHandler))
		require.False(t, *called, "unknown method")
		msg.WithCalleeMethod("s_a_m")
		require.Nil(t, filters(ctx, nil, nil, noopClientHandler))
		require.True(t, *called)
	}
	testClientFilters(clientFilters, &cmfCalled)

	noopServerHandler := func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil }
	testServerFilters := func(filters filter.ServerFilter, called *bool) {
		ctx := trpc.BackgroundContext()
		msg := trpc.Message(ctx)
		msg.WithCalleeServiceName("unknown-server-service")
		_, err := filters(ctx, nil, noopServerHandler)
		require.Nil(t, err)
		require.False(t, *called, "unknown service")

		msg.WithCalleeServiceName("s_a")
		_, err = filters(ctx, nil, noopServerHandler)
		require.Nil(t, err)
		require.False(t, *called, "missing method")

		msg.WithCalleeMethod("unknown-method")
		_, err = filters(ctx, nil, noopServerHandler)
		require.Nil(t, err)
		require.False(t, *called, "unknown method")

		msg.WithCalleeMethod("s_a_m")
		_, err = filters(ctx, nil, noopServerHandler)
		require.Nil(t, err)
		require.True(t, *called)
	}
	testServerFilters(serverFilters, &smfCalled)
}

const yamlCfg = `
server:
  - name: s_a
    methods:
      - name: s_a_m
        filters: [smf]

client:
  - name: s_a
    methods:
      - name: s_a_m
        filters: [cmf]
`
