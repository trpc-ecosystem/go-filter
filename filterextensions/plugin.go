// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package filterextensions 是对 tRPC-Go yaml 配置的一个拓展。它允许用户定义 method 粒度的 filter。
package filterextensions

import (
	"context"
	"fmt"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	// PluginType plugin 的类型
	PluginType = "filter_extensions"
	// PluginName plugin 的名字
	PluginName = "method_filters"
	// MethodFilters plugin 注册的 filter 的名字
	MethodFilters = "method_filters"
)

func init() {
	plugin.Register(PluginName, &serviceMethodFiltersPlugin{})
}

type methodClientFilters map[string][]filter.ClientFilter
type methodServerFilters map[string][]filter.ServerFilter
type serviceMethodClientFilters map[string]map[string][]filter.ClientFilter
type serviceMethodServerFilters map[string]map[string][]filter.ServerFilter

type serviceMethodFiltersPlugin struct {
	client serviceMethodClientFilters
	server serviceMethodServerFilters
}

// Type 返回插件类型。
func (p *serviceMethodFiltersPlugin) Type() string {
	return PluginType
}

// Setup 初始化插件。
func (p *serviceMethodFiltersPlugin) Setup(_ string, dec plugin.Decoder) error {
	var cfg cfg
	if err := dec.Decode(&cfg); err != nil {
		return err
	}

	// 从 tRPC 全局注册的 filter 中寻找，没找到则报错。
	client, err := loadClientFilters(cfg.Client, filter.GetClient)
	if err != nil {
		return fmt.Errorf("failed to load client service method filters, err: %w", err)
	}
	p.client = client
	server, err := loadServerFilters(cfg.Server, filter.GetServer)
	if err != nil {
		return fmt.Errorf("failed to load server service method filters, err: %w", err)
	}
	p.server = server
	filter.Register(MethodFilters, newServerIntercept(p.server), newClientIntercept(p.client))
	return nil
}

func newClientIntercept(
	serviceFilters serviceMethodClientFilters,
) filter.ClientFilter {
	return func(ctx context.Context, req, rsp interface{}, handler filter.ClientHandleFunc) error {
		msg := trpc.Message(ctx)
		service := msg.CalleeServiceName()
		method := msg.CalleeMethod()
		if methodFilters, ok := serviceFilters[service]; ok {
			if filters, ok := methodFilters[method]; ok {
				return filter.ClientChain(filters).Filter(ctx, req, rsp, handler)
			}
		}
		return handler(ctx, req, rsp)
	}
}

func newServerIntercept(
	serviceFilters serviceMethodServerFilters,
) filter.ServerFilter {
	return func(ctx context.Context, req interface{}, handler filter.ServerHandleFunc) (interface{}, error) {
		msg := trpc.Message(ctx)
		service := msg.CalleeServiceName()
		method := msg.CalleeMethod()
		if methodFilters, ok := serviceFilters[service]; ok {
			if filters, ok := methodFilters[method]; ok {
				return filter.ServerChain(filters).Filter(ctx, req, handler)
			}
		}
		return handler(ctx, req)
	}

}

func loadClientFilters(
	services []cfgService,
	filterLoader func(name string) filter.ClientFilter,
) (serviceMethodClientFilters, error) {
	loadMethodFilters := func(filterNames []string) ([]filter.ClientFilter, error) {
		filters := make([]filter.ClientFilter, 0, len(filterNames))
		for _, filterName := range filterNames {
			f := filterLoader(filterName)
			if f == nil {
				return nil, fmt.Errorf("filter %s not registered", filterName)
			}
			filters = append(filters, f)
		}
		return filters, nil
	}

	smf := make(serviceMethodClientFilters, len(services))
	for _, service := range services {
		mf := make(methodClientFilters, len(service.Methods))
		for _, method := range service.Methods {
			f, err := loadMethodFilters(method.Filters)
			if err != nil {
				return nil, err
			}
			mf[method.Name] = f
		}
		smf[service.Name] = mf
	}
	return smf, nil
}

func loadServerFilters(
	services []cfgService,
	filterLoader func(name string) filter.ServerFilter,
) (serviceMethodServerFilters, error) {
	loadMethodFilters := func(filterNames []string) ([]filter.ServerFilter, error) {
		filters := make([]filter.ServerFilter, 0, len(filterNames))
		for _, filterName := range filterNames {
			f := filterLoader(filterName)
			if f == nil {
				return nil, fmt.Errorf("filter %s not registered", filterName)
			}
			filters = append(filters, f)
		}
		return filters, nil
	}

	smf := make(serviceMethodServerFilters, len(services))
	for _, service := range services {
		mf := make(methodServerFilters, len(service.Methods))
		for _, method := range service.Methods {
			f, err := loadMethodFilters(method.Filters)
			if err != nil {
				return nil, err
			}
			mf[method.Name] = f
		}
		smf[service.Name] = mf
	}
	return smf, nil
}
