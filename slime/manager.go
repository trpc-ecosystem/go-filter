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
	"context"
	"fmt"

	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/plugin"

	"trpc.group/trpc-go/trpc-filter/slime/hedging"
	"trpc.group/trpc-go/trpc-filter/slime/once"
	"trpc.group/trpc-go/trpc-filter/slime/retry"
	"trpc.group/trpc-go/trpc-filter/slime/throttle"
)

const (
	pluginType = "slime"
	pluginName = "default"
	filterName = "slime"
)

func init() {
	plugin.Register(pluginName, defaultManager)
}

// retryHedging is an abstraction of retry, hedging or once.
type retryHedging interface {
	Invoke(ctx context.Context, req, rsp interface{}, f filter.ClientHandleFunc) (err error)
}

// This package prefer a singleton retryHedgingManager.
var defaultManager = &retryHedgingManager{}

// retryHedgingManager implements plugin.Factory.
type retryHedgingManager struct {
	noRetryHedging *once.Once
	retries        map[string]*retry.Retry
	hedges         map[string]*hedging.Hedging
	services       map[string]*service
}

type service struct {
	retryHedging retryHedging
	methods      map[string]retryHedging
	throttle     throttle.Throttler
}

// Type implements Type in plugin.Factory.
func (p *retryHedgingManager) Type() string {
	return pluginType
}

// Setup implements Setup in plugin.Factory.
// It also register the retry/hedging filter.
func (p *retryHedgingManager) Setup(_ string, configDec plugin.Decoder) error {
	clientCfg := clientCfg{}
	err := configDec.Decode(&clientCfg)
	if err != nil {
		return fmt.Errorf("failed to parse retry/hedging cfg, err: %w", err)
	}

	p.noRetryHedging = once.New()
	p.retries = make(map[string]*retry.Retry)
	p.hedges = make(map[string]*hedging.Hedging)
	p.services = make(map[string]*service)

	for _, sCfg := range clientCfg.Services {
		sCfg.repair()

		s, ok := p.services[sCfg.Name]
		if !ok {
			s = &service{
				methods: make(map[string]retryHedging),
			}
			throt, err := newThrottle(sCfg.Throttle)
			if err != nil {
				return err
			}
			s.throttle = throt
			p.services[sCfg.Name] = s
		}

		s.retryHedging, err = p.newRetryHedging(sCfg.RetryHedging, s.throttle)
		if err != nil {
			return err
		}

		for _, mCfg := range sCfg.Methods {
			if mCfg.RetryHedging == nil {
				continue
			}
			rh, err := p.newRetryHedging(*mCfg.RetryHedging, s.throttle)
			if err != nil {
				return err
			}
			s.methods[mCfg.Callee] = rh
		}
	}

	filter.Register(filterName, nil, interceptor)

	return nil
}

// newRetryHedging return retryHedging from retryHedgingCfg.
// If no retry or hedging cfg is found, the trivial noRetryHedging will be returned.
// When both retry and hedging are provided, retry has higher priority.
func (p *retryHedgingManager) newRetryHedging(cfg retryHedgingCfg, throt throttle.Throttler) (retryHedging, error) {
	if cfg.Retry != nil {
		cfg.Retry.repair()
		r, err := p.newRetry(*cfg.Retry)
		if err != nil {
			return nil, err
		}
		return r.NewThrottledRetry(throt), nil
	}
	if cfg.Hedging != nil {
		cfg.Hedging.repair()
		h, err := p.newHedging(*cfg.Hedging)
		if err != nil {
			return nil, err
		}
		return h.NewThrottledHedging(throt), nil
	}

	return p.noRetryHedging, nil
}

// newRetry create a new retry if it does not exist.
func (p *retryHedgingManager) newRetry(cfg retryCfg) (*retry.Retry, error) {
	if r, ok := p.retries[cfg.Name]; ok {
		return r, nil
	}

	var opts []retry.Opt
	if bf := cfg.Backoff.Linear; bf != nil {
		opts = append(opts, retry.WithLinearBackoff(bf...))
	}
	if bf := cfg.Backoff.Exponential; bf != nil {
		opts = append(opts, retry.WithExpBackoff(bf.Initial, bf.Maximum, bf.Multiplier))
	}
	if cfg.SkipVisitedNodes != nil {
		opts = append(opts, retry.WithSkipVisitedNodes(*cfg.SkipVisitedNodes))
	}

	r, err := retry.New(cfg.MaxAttempts, cfg.RetryableECs, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create retry policy, err: %w", err)
	}

	p.retries[cfg.Name] = r
	return r, nil
}

// newHedging create a new hedging if it does not exist.
func (p *retryHedgingManager) newHedging(cfg hedgingCfg) (*hedging.Hedging, error) {
	if h, ok := p.hedges[cfg.Name]; ok {
		return h, nil
	}

	opts := []hedging.Opt{hedging.WithStaticHedgingDelay(cfg.HedgingDelay)}
	if cfg.SkipVisitedNodes != nil {
		opts = append(opts, hedging.WithSkipVisitedNodes(*cfg.SkipVisitedNodes))
	}

	h, err := hedging.New(cfg.MaxAttempts, cfg.NonFatalECs, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create hedging policy, err: %w", err)
	}

	p.hedges[cfg.Name] = h
	return h, nil
}

const (
	defaultMaxTokens  = 10
	defaultTokenRatio = 0.1
)

// newThrottle create a new throttle.Throttler from throttleCfg.
func newThrottle(cfg *throttleCfg) (throttle.Throttler, error) {
	// Throttle is not explicitly configured, enable default throttle.
	if cfg == nil {
		return throttle.NewTokenBucket(defaultMaxTokens, defaultTokenRatio)
	}

	// Throttle is explicitly configured as empty. Use Noop to disable it.
	if cfg.MaxTokens == 0 && cfg.TokenRatio == 0 {
		return throttle.NewNoop(), nil
	}

	// Use default values for unconfigured fields in throttle.
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	if cfg.TokenRatio == 0 {
		cfg.TokenRatio = defaultTokenRatio
	}

	return throttle.NewTokenBucket(cfg.MaxTokens, cfg.TokenRatio)
}
