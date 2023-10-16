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

package slime

import (
	"time"

	"trpc.group/trpc-go/trpc-go/errs"

	"github.com/google/uuid"
)

// clientCfg is the configuration of slime client.
type clientCfg struct {
	Services []serviceCfg `yaml:"service"`
}

// serviceCfg is the configuration of slime service.
type serviceCfg struct {
	Name         string          `yaml:"name"`
	Callee       string          `yaml:"callee"`
	Throttle     *throttleCfg    `yaml:"retry_hedging_throttle"`
	RetryHedging retryHedgingCfg `yaml:"retry_hedging"`
	Methods      []methodCfg     `yaml:"methods"`
}

// repair is used to make configuration file consistent with tRPC-Go.
// Slime only needs naming service, aka Name, and does not care about proto service, aka Callee.
func (cfg *serviceCfg) repair() {
	if cfg.Name == "" {
		cfg.Name = cfg.Callee
	}
}

// throttleCfg is the configuration of slime throttle.
type throttleCfg struct {
	MaxTokens  float64 `yaml:"max_tokens"`
	TokenRatio float64 `yaml:"token_ratio"`
}

// methodCfg is the configuration of slime method.
type methodCfg struct {
	Callee       string           `yaml:"callee"`
	RetryHedging *retryHedgingCfg `yaml:"retry_hedging"`
}

// retryHedgingCfg is the configuration of slime retry and hedging.
type retryHedgingCfg struct {
	Retry   *retryCfg   `yaml:"retry"`
	Hedging *hedgingCfg `yaml:"hedging"`
}

// hedgingCfg is the configuration of slime hedging.
type hedgingCfg struct {
	Name             string        `yaml:"name"`
	MaxAttempts      int           `yaml:"max_attempts"`
	HedgingDelay     time.Duration `yaml:"hedging_delay"`
	NonFatalECs      []int         `yaml:"non_fatal_error_codes"`
	SkipVisitedNodes *bool         `yaml:"skip_visited_nodes"`
}

var (
	defaultHedgingMaxAttempts = 2
	defaultNonFatalECs        = []int{
		int(errs.RetServerTimeout),
		int(errs.RetClientConnectFail),
		int(errs.RetClientRouteErr),
		int(errs.RetClientNetErr),
	}
)

// repair fix hedgingCfg With default values.
func (cfg *hedgingCfg) repair() {
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = defaultHedgingMaxAttempts
	}
	if cfg.Name == "" {
		cfg.Name = "hedging-" + uuid.New().String()
	}
	if len(cfg.NonFatalECs) == 0 {
		cfg.NonFatalECs = defaultNonFatalECs
	}
}

// retryCfg is the configuration of slime retry.
type retryCfg struct {
	Name             string     `yaml:"name"`
	MaxAttempts      int        `yaml:"max_attempts"`
	Backoff          backoffCfg `yaml:"backoff"`
	RetryableECs     []int      `yaml:"retryable_error_codes"`
	SkipVisitedNodes *bool      `yaml:"skip_visited_nodes"`
}

var (
	defaultRetryMaxAttempts = 2
	defaultRetryableECs     = []int{
		int(errs.RetServerTimeout),
		int(errs.RetClientConnectFail),
		int(errs.RetClientRouteErr),
		int(errs.RetClientNetErr),
	}
)

// repair fix retryCfg with default values.
func (cfg *retryCfg) repair() {
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = defaultRetryMaxAttempts
	}
	if cfg.Name == "" {
		cfg.Name = "retry-" + uuid.New().String()
	}
	if len(cfg.RetryableECs) == 0 {
		cfg.RetryableECs = defaultRetryableECs
	}
}

// backoffCfg is the configuration of slime backoff.
type backoffCfg struct {
	Exponential *struct {
		Initial    time.Duration `yaml:"initial"`
		Maximum    time.Duration `yaml:"maximum"`
		Multiplier int           `yaml:"multiplier"`
	} `yaml:"exponential"`
	Linear []time.Duration `yaml:"linear"`
}
