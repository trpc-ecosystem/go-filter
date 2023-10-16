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

// Package degrade 熔断限流插件
// 熔断限流插件
// author:boxbai@tencent.com
// date 2019-11-11
package degrade

import (
	"context"
	"math/rand"
	"time"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const (
	pluginType          = "circuitbreaker"
	pluginName          = "degrade"
	infoDegradeRateZero = "because the derade_rate is zero,so exit plugin"
	infoDegradeRate100  = "the derade_rate is 100, so exit plugin"
	errDegardeReturn    = "service is degrade..."
	systemDegradeErrNo  = 22
)

var (
	isDegrade bool
	cfg       Config
	sema      chan struct{}
)

// Config 熔断配置结构体声明
type Config struct {
	// Load5 5 分钟 触发熔断的阈值，恢复时使用实时 load1 <= 本值来判断，更敏感
	Load5 float64 `yaml:"load5"`
	// CPUIdle cpuidle 触发熔断的阈值
	CPUIdle int `yaml:"cpu_idle"`
	// MemoryUsepercent 内存使用率百分比 触发熔断的阈值
	MemoryUsePercent int `yaml:"memory_use_p"`
	// DegradeRate 流量抛弃比例，目前使用随机算法抛弃，迭代加入其他均衡算法
	DegradeRate int `yaml:"degrade_rate"`
	// Interval 心跳时间间隔，主要控制多久更新一次熔断开关状态
	Interval int `yaml:"interval"`
	// Modulename 模块名，后续用于上报鹰眼或其他日志
	Modulename string `yaml:"modulename"`
	// Whitelist 白名单，用于后续跳过不被熔断控制的业务接口
	Whitelist string `yaml:"whitelist"`
	// IsActive 标志熔断是否生效
	IsActive bool `yaml:"-"`
	// MaxConcurrentCnt 最大并发请求数，<=0 时不开启，控制最大并发请求数，和上述熔断互为补充，能防止突发流量把服务打死，比如 1ms 内突然进入 100W 请求
	MaxConcurrentCnt int `yaml:"max_concurrent_cnt"`
	// MaxTimeOutMs 超过最大并发请求数时，最多等待 MaxTimeOutMs 才决定是熔断还是继续处理
	MaxTimeOutMs int `yaml:"max_timeout_ms"`
}

// Degrade 熔断插件默认初始化
type Degrade struct{}

// Filter 声明熔断组件的 filter 来充当拦截器
func Filter(
	ctx context.Context, req interface{}, handler filter.ServerHandleFunc,
) (interface{}, error) {
	if isDegrade {
		randNum := rand.Intn(100)
		if randNum >= cfg.DegradeRate {
			return nil, errs.New(systemDegradeErrNo, errDegardeReturn)
		}

	}
	if enableConcurrency() {
		select {
		// 未达到最大并发请求数
		case sema <- struct{}{}:
			defer func() {
				<-sema
			}()
		// 达到最大并发请求数，直接丢弃请求
		default:
			return nil, errs.New(systemDegradeErrNo, errDegardeReturn)
		}
	}

	return handler(ctx, req)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	plugin.Register(pluginName, &Degrade{})
}

// Type 返回插件类型
func (p *Degrade) Type() string {
	return pluginType
}

// Setup 注册
func (p *Degrade) Setup(name string, decoder plugin.Decoder) error {
	if err := decoder.Decode(&cfg); err != nil {
		return err
	}
	if cfg.DegradeRate == 0 {
		log.Info(infoDegradeRateZero)
		return nil
	}
	if cfg.DegradeRate == 100 {
		log.Info(infoDegradeRate100)
		return nil
	}
	if cfg.Interval == 0 {
		cfg.Interval = 60
	}
	if enableConcurrency() {
		sema = make(chan struct{}, cfg.MaxConcurrentCnt)
	}
	go UpdateSysInfoPerTime()
	go func() {
		var load1, load5 float64
		for range time.Tick(time.Duration(cfg.Interval) * time.Second) {
			cpuIdle := GetCPUIdle()
			mem := int(GetMemoryStat())
			load, err := GetLoadAvg()
			if err != nil {
				load5 = 0
				load1 = 0
			} else {
				load1 = load.Load1
				load5 = load.Load5
			}
			if load5 > cfg.Load5 || mem > cfg.MemoryUsePercent || cpuIdle < cfg.CPUIdle {
				isDegrade = true
			}
			if load1 <= cfg.Load5 && mem <= cfg.MemoryUsePercent && cpuIdle >= cfg.CPUIdle {
				isDegrade = false
			}
			log.Infof("%s cpu_idle:%d mem_usage:%d load5:%f,degrade:%t",
				time.Now(), cpuIdle, mem, load5, isDegrade)
		}
	}()

	filter.Register(pluginName, Filter, nil)

	return nil
}
