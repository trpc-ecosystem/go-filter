// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package slime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trpc.group/trpc-go/trpc-filter/slime/hedging"
	"trpc.group/trpc-go/trpc-filter/slime/retry"
	"trpc.group/trpc-go/trpc-filter/slime/view"
	"trpc.group/trpc-go/trpc-filter/slime/view/log"
	"trpc.group/trpc-go/trpc-filter/slime/view/metrics"
)

// SetHedgingDynamicDelay sets the dynamic delay function for hedging policy name.
func SetHedgingDynamicDelay(name string, dynamicDelay func() time.Duration) error {
	if dynamicDelay == nil {
		return errors.New("hedging dynamicDelay must be non-nil")
	}

	h, ok := defaultManager.hedges[name]
	if !ok {
		return fmt.Errorf("hedging policy %s is not found", name)
	}

	hedging.WithDynamicHedgingDelay(dynamicDelay)(h)
	return nil
}

// SetAllHedgingDynamicDelay sets the dynamic delay function for all hedging policies.
func SetAllHedgingDynamicDelay(dynamicDelay func() time.Duration) error {
	for name := range defaultManager.hedges {
		if err := SetHedgingDynamicDelay(name, dynamicDelay); err != nil {
			return err
		}
	}
	return nil
}

// SetHedgingNonFatalError sets the nonFatalErr function for hedging policy name.
func SetHedgingNonFatalError(name string, nonFatalErr func(error) bool) error {
	if nonFatalErr == nil {
		return errors.New("hedging nonFatalErr must be non-nil")
	}

	h, ok := defaultManager.hedges[name]
	if !ok {
		return fmt.Errorf("hedgign policy %s is not found", name)
	}

	hedging.WithNonFatalErr(nonFatalErr)(h)
	return nil
}

// SetAllHedgingNonFatalError sets the nonFatalErr functions for all hedging policies.
func SetAllHedgingNonFatalError(nonFatalErr func(error) bool) error {
	for name := range defaultManager.hedges {
		if err := SetHedgingNonFatalError(name, nonFatalErr); err != nil {
			return err
		}
	}
	return nil
}

// SetHedgingRspToErr sets the rsp body error convert function for hedging policy name.
func SetHedgingRspToErr(name string, rspToErr func(interface{}) error) error {
	if rspToErr == nil {
		return errors.New("hedging rspToErr must be non-nil")
	}
	h, ok := defaultManager.hedges[name]
	if !ok {
		return fmt.Errorf("hedgign policy %s is not found", name)
	}
	hedging.WithRspToErr(rspToErr)(h)
	return nil
}

// SetAllHedgingRspToErr sets the rspToErr functions for all hedging policies.
func SetAllHedgingRspToErr(rspToErr func(interface{}) error) error {
	for name := range defaultManager.hedges {
		if err := SetHedgingRspToErr(name, rspToErr); err != nil {
			return err
		}
	}
	return nil
}

// SetHedgingConditionalLog sets the conditional log for the hedging policy.
func SetHedgingConditionalLog(name string, l log.Logger, condition func(view.Stat) bool) error {
	if l == nil {
		return errors.New("logger must be non-nil")
	}
	if condition == nil {
		return errors.New("condition must be non-nil")
	}

	h, ok := defaultManager.hedges[name]
	if !ok {
		return fmt.Errorf("hedging policy %s is not found", name)
	}

	hedging.WithConditionalLog(l, condition)(h)
	return nil
}

// SetHedgingConditionalCtxLog sets the conditional log for the hedging policy.
func SetHedgingConditionalCtxLog(name string, l log.CtxLogger, condition func(view.Stat) bool) error {
	if l == nil {
		return errors.New("logger must be non-nil")
	}
	if condition == nil {
		return errors.New("condition must be non-nil")
	}

	h, ok := defaultManager.hedges[name]
	if !ok {
		return fmt.Errorf("hedging policy %s is not found", name)
	}

	hedging.WithConditionalCtxLog(l, condition)(h)
	return nil
}

// SetAllHedgingConditionalLog sets the conditional log for all hedging polices.
func SetAllHedgingConditionalLog(l log.Logger, condition func(view.Stat) bool) error {
	for name := range defaultManager.hedges {
		if err := SetHedgingConditionalLog(name, l, condition); err != nil {
			return err
		}
	}
	return nil
}

// SetAllHedgingConditionalCtxLog sets the conditional log for all hedging polices.
func SetAllHedgingConditionalCtxLog(l log.CtxLogger, condition func(view.Stat) bool) error {
	for name := range defaultManager.hedges {
		if err := SetHedgingConditionalCtxLog(name, l, condition); err != nil {
			return err
		}
	}
	return nil
}

// SetHedgingEmitter sets the emitter for the hedging policy.
func SetHedgingEmitter(name string, emitter metrics.Emitter) error {
	if emitter == nil {
		return errors.New("nil emitter")
	}

	h, ok := defaultManager.hedges[name]
	if !ok {
		return fmt.Errorf("hedging policy %s is not found", name)
	}

	hedging.WithEmitter(emitter)(h)
	return nil
}

// SetAllHedgingEmitter sets the emitter for all hedging polices.
func SetAllHedgingEmitter(emitter metrics.Emitter) error {
	for name := range defaultManager.hedges {
		if err := SetHedgingEmitter(name, emitter); err != nil {
			return err
		}
	}
	return nil
}

// SetRetryBackoff sets the backoff function for retry policy name.
func SetRetryBackoff(name string, backoff func(attempt int) time.Duration) error {
	if backoff == nil {
		return errors.New("retry backoff must be non-nil")
	}

	r, ok := defaultManager.retries[name]
	if !ok {
		return fmt.Errorf("retry policy %s is not found", name)
	}

	return retry.WithBackoff(backoff)(r)
}

// SetAllRetryBackoff sets the backoff function for all retry polices.
func SetAllRetryBackoff(backoff func(attempt int) time.Duration) error {
	for name := range defaultManager.retries {
		if err := SetRetryBackoff(name, backoff); err != nil {
			return err
		}
	}
	return nil
}

// SetRetryRetryableErr sets the retryable error function for retry policy name.
func SetRetryRetryableErr(name string, retryableErr func(error) bool) error {
	if retryableErr == nil {
		return errors.New("retry retryableErr must be non-nil")
	}

	r, ok := defaultManager.retries[name]
	if !ok {
		return fmt.Errorf("retry policy %s is not found", name)
	}

	return retry.WithRetryableErr(retryableErr)(r)
}

// SetAllRetryRetryableErr sets the retryable error function for all retry polices.
func SetAllRetryRetryableErr(retryableErr func(error) bool) error {
	for name := range defaultManager.retries {
		if err := SetRetryRetryableErr(name, retryableErr); err != nil {
			return err
		}
	}
	return nil
}

// SetRetryRspToErr sets the rsp body error convert function for retry policy name.
func SetRetryRspToErr(name string, rspToErr func(interface{}) error) error {
	if rspToErr == nil {
		return errors.New("retry rspToErr must be non-nil")
	}
	r, ok := defaultManager.retries[name]
	if !ok {
		return fmt.Errorf("retry policy %s is not found", name)
	}
	return retry.WithRspToErr(rspToErr)(r)
}

// SetAllRetryRspToErr sets the rsp body error convert function for all retry polices.
func SetAllRetryRspToErr(rspToErr func(interface{}) error) error {
	for name := range defaultManager.retries {
		if err := SetRetryRspToErr(name, rspToErr); err != nil {
			return err
		}
	}
	return nil
}

// SetRetryConditionalLog sets the conditional log for the retry policy.
func SetRetryConditionalLog(name string, l log.Logger, condition func(view.Stat) bool) error {
	if l == nil {
		return errors.New("logger must be non-nil")
	}
	if condition == nil {
		return errors.New("condition must be non-nil")
	}

	r, ok := defaultManager.retries[name]
	if !ok {
		return fmt.Errorf("retry policy %s is not found", name)
	}

	return retry.WithConditionalLog(l, condition)(r)
}

// SetRetryConditionalCtxLog sets the conditional log for the retry policy.
func SetRetryConditionalCtxLog(name string, l log.CtxLogger, condition func(view.Stat) bool) error {
	if l == nil {
		return errors.New("logger must be non-nil")
	}
	if condition == nil {
		return errors.New("condition must be non-nil")
	}

	r, ok := defaultManager.retries[name]
	if !ok {
		return fmt.Errorf("retry policy %s is not found", name)
	}

	return retry.WithConditionalCtxLog(l, condition)(r)
}

// SetAllRetryConditionalLog sets the conditional log for all retry polices.
func SetAllRetryConditionalLog(l log.Logger, condition func(view.Stat) bool) error {
	for name := range defaultManager.retries {
		if err := SetRetryConditionalLog(name, l, condition); err != nil {
			return err
		}
	}
	return nil
}

// SetAllRetryConditionalCtxLog sets the conditional log for all retry polices.
func SetAllRetryConditionalCtxLog(l log.CtxLogger, condition func(view.Stat) bool) error {
	for name := range defaultManager.retries {
		if err := SetRetryConditionalCtxLog(name, l, condition); err != nil {
			return err
		}
	}
	return nil
}

// SetRetryEmitter sets the emitter for the retry policy.
func SetRetryEmitter(name string, emitter metrics.Emitter) error {
	if emitter == nil {
		return errors.New("nil emitter")
	}

	r, ok := defaultManager.retries[name]
	if !ok {
		return fmt.Errorf("retry policy %s is not found", name)
	}

	return retry.WithEmitter(emitter)(r)
}

// SetAllRetryEmitter sets the emitter for all retry polices.
func SetAllRetryEmitter(emitter metrics.Emitter) error {
	for name := range defaultManager.retries {
		if err := SetRetryEmitter(name, emitter); err != nil {
			return err
		}
	}
	return nil
}

type disabledKey struct{}
type disabledVal struct{}

// WithDisabled returns a context which skips retry or hedging.
func WithDisabled(ctx context.Context) context.Context {
	return context.WithValue(ctx, disabledKey{}, disabledVal{})
}

func disabled(ctx context.Context) bool {
	_, ok := ctx.Value(disabledKey{}).(disabledVal)
	return ok
}
