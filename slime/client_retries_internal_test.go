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

package slime

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"trpc.group/trpc-go/trpc-filter/slime/hedging"
	"trpc.group/trpc-go/trpc-filter/slime/retry"
	"trpc.group/trpc-go/trpc-filter/slime/throttle"
	"trpc.group/trpc-go/trpc-filter/slime/view"
	"trpc.group/trpc-go/trpc-filter/slime/view/metrics"
	"trpc.group/trpc-go/trpc-go"
)

const cfgFile = "client_retries_test.yaml"

const (
	serviceWelcome     = "trpc.app.server.Welcome"
	serviceGreeting    = "trpc.app.server.Greeting"
	serviceUnspecified = "trpc.app.server.Unspecified"

	methodHello       = "Hello"
	methodHi          = "Hi"
	methodGreet       = "Greet"
	methodYo          = "Yo"
	methodUnspecified = "Unspecified"
)

func TestPluginsType(t *testing.T) {
	require.Equal(t, pluginType, defaultManager.Type())
}

func TestRetryHedgingManagerSetup(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.Equal(t, 2, len(defaultManager.retries))
	require.Equal(t, 2, len(defaultManager.hedges))
	require.Equal(t, 2, len(defaultManager.services))

	r1, ok := defaultManager.retries["retry1"]
	require.True(t, ok)
	r2, ok := defaultManager.retries["retry2"]
	require.True(t, ok)
	h1, ok := defaultManager.hedges["hedging1"]
	require.True(t, ok)
	h2, ok := defaultManager.hedges["hedging2"]
	require.True(t, ok)

	service, ok := defaultManager.services[serviceWelcome]
	require.True(t, ok)
	tr, ok := service.retryHedging.(*retry.ThrottledRetry)
	require.True(t, ok)
	require.Equal(t, r1, tr.Retry)

	_, ok = service.throttle.(*throttle.TokenBucket)
	require.True(t, ok)

	method, ok := service.methods[methodHello]
	require.True(t, ok)
	tr, ok = method.(*retry.ThrottledRetry)
	require.True(t, ok)
	require.Equal(t, r2, tr.Retry)

	method, ok = service.methods[methodHi]
	require.True(t, ok)
	th, ok := method.(*hedging.ThrottledHedging)
	require.True(t, ok)
	require.Equal(t, h1, th.Hedging)

	method, ok = service.methods[methodGreet]
	require.True(t, ok)
	require.Equal(t, defaultManager.noRetryHedging, method)

	_, ok = service.methods[methodYo]
	require.False(t, ok)

	service, ok = defaultManager.services[serviceGreeting]
	require.True(t, ok)
	th, ok = service.retryHedging.(*hedging.ThrottledHedging)
	require.True(t, ok)
	require.Equal(t, h2, th.Hedging)

	_, ok = service.throttle.(*throttle.TokenBucket)
	require.True(t, ok)
}

func TestFilter(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	r1, ok := defaultManager.retries["retry1"]
	require.True(t, ok)
	r2, ok := defaultManager.retries["retry2"]
	require.True(t, ok)
	h1, ok := defaultManager.hedges["hedging1"]
	require.True(t, ok)
	h2, ok := defaultManager.hedges["hedging2"]
	require.True(t, ok)

	ctx := trpc.BackgroundContext()
	msg := trpc.Message(ctx)

	rh := getRetryHedging(ctx)
	require.Nil(t, rh)

	msg.WithCalleeServiceName(serviceWelcome)
	rh = getRetryHedging(ctx)
	tr, ok := rh.(*retry.ThrottledRetry)
	require.True(t, ok, "msg missing method should use rh of service")
	require.Equal(t, r1, tr.Retry, "msg missing method should use rh of service")

	msg.WithCalleeMethod(methodHello)
	rh = getRetryHedging(ctx)
	tr, ok = rh.(*retry.ThrottledRetry)
	require.True(t, ok)
	require.Equal(t, r2, tr.Retry)

	msg.WithCalleeMethod(methodHi)
	rh = getRetryHedging(ctx)
	th, ok := rh.(*hedging.ThrottledHedging)
	require.True(t, ok)
	require.Equal(t, h1, th.Hedging)

	msg.WithCalleeMethod(methodGreet)
	rh = getRetryHedging(ctx)
	require.Equal(t, defaultManager.noRetryHedging, rh)

	msg.WithCalleeMethod(methodYo)
	rh = getRetryHedging(ctx)
	tr, ok = rh.(*retry.ThrottledRetry)
	require.True(t, ok)
	require.Equal(t, r1, tr.Retry)

	msg.WithCalleeMethod(methodUnspecified)
	rh = getRetryHedging(ctx)
	tr, ok = rh.(*retry.ThrottledRetry)
	require.True(t, ok)
	require.Equal(t, r1, tr.Retry, "should use default retryHedging of service")

	msg.WithCalleeServiceName(serviceGreeting)
	rh = getRetryHedging(ctx)
	th, ok = rh.(*hedging.ThrottledHedging)
	require.True(t, ok)
	require.Equal(t, h2, th.Hedging)

	msg.WithCalleeServiceName(serviceUnspecified)
	rh = getRetryHedging(ctx)
	require.Nil(t, rh)
}

func TestSetHedgingDynamicDelay(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetHedgingDynamicDelay("hedging3", func() time.Duration { return time.Second })
	require.NotNil(t, err, "hedging3 should not exist")

	err = SetHedgingDynamicDelay("hedging2", func() time.Duration { return time.Second })
	require.Nil(t, err)
}

func TestSetAllHedgingDynamicDelay(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllHedgingDynamicDelay(nil))
	require.Nil(t, SetAllHedgingDynamicDelay(func() time.Duration {
		return time.Second
	}))
}

func TestSetHedgingNonFatalError(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetHedgingNonFatalError("hedging3", func(err error) bool { return true })
	require.NotNil(t, err, "hedging3 should not exist")

	err = SetHedgingNonFatalError("hedging2", func(err error) bool { return true })
	require.Nil(t, err)
}

func TestSetAllHedgingNonFatalError(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllHedgingNonFatalError(nil))
	require.Nil(t, SetAllHedgingNonFatalError(func(err error) bool {
		return true
	}))
}

func TestSetHedgingConditionalLog(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetHedgingConditionalLog("hedging3", &ConsoleLog{}, func(view.Stat) bool { return false }),
		"hedging3 should not exist")
	require.Nil(t, SetHedgingConditionalLog("hedging2", &ConsoleLog{}, func(view.Stat) bool { return true }))

	require.NotNil(t, SetHedgingConditionalCtxLog("hedging2", nil, func(view.Stat) bool {
		return true
	}), "log cannot be nil")
	require.NotNil(t, SetHedgingConditionalCtxLog("hedging2", &CtxConsoleLog{}, nil),
		"condition cannot be nil")
	require.NotNil(t, SetHedgingConditionalCtxLog("hedging3", &CtxConsoleLog{}, func(view.Stat) bool {
		return true
	}), "hedging3 should not exist")
	require.Nil(t, SetHedgingConditionalCtxLog("hedging2", &CtxConsoleLog{}, func(view.Stat) bool {
		return true
	}))
}

func TestSetAllHedgingConditionalLog(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllHedgingConditionalLog(nil, nil))
	require.NotNil(t, SetAllHedgingConditionalLog(&ConsoleLog{}, nil))
	require.Nil(t, SetAllHedgingConditionalLog(&ConsoleLog{}, func(view.Stat) bool {
		return true
	}))
	require.Nil(t, SetAllHedgingConditionalCtxLog(&CtxConsoleLog{}, func(view.Stat) bool {
		return true
	}))
}

func TestSetHedgingEmitter(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetHedgingEmitter("hedging3", metrics.Noop{})
	require.NotNil(t, err, "hedging3 should not exist")

	err = SetHedgingEmitter("hedging2", nil)
	require.NotNil(t, err, "emitter should not be nil")

	err = SetHedgingEmitter("hedging2", metrics.Noop{})
	require.Nil(t, err)
}

func TestSetAllHedgingEmitter(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllHedgingEmitter(nil))
	require.Nil(t, SetAllHedgingEmitter(metrics.Noop{}))
}

func TestSetHedgingRspToErr(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetHedgingRspToErr("hedging3", func(rsp interface{}) error { return nil })
	require.NotNil(t, err, "retry3 should not exist")

	err = SetHedgingRspToErr("hedging2", func(rsp interface{}) error { return nil })
	require.Nil(t, err)
}

func TestSetAllHedgingRspToErr(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllHedgingRspToErr(nil))
	require.Nil(t, SetAllHedgingRspToErr(func(rsp interface{}) error {
		return nil
	}))
}

func TestSetRetryBackoff(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetRetryBackoff("retry3", func(attempt int) time.Duration { return time.Second })
	require.NotNil(t, err, "retry3 should not exist")

	err = SetRetryBackoff("retry2", func(attempt int) time.Duration { return time.Second })
	require.Nil(t, err)
}

func TestSetAllRetryBackoff(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllRetryBackoff(nil))
	require.Nil(t, SetAllRetryBackoff(func(attempt int) time.Duration {
		return time.Second
	}))
}

func TestSetRetryRetryableErr(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetRetryRetryableErr("retry3", func(err error) bool { return true })
	require.NotNil(t, err, "retry3 should not exist")

	err = SetRetryRetryableErr("retry2", func(err error) bool { return true })
	require.Nil(t, err)
}

func TestSetAllRetryRetryableErr(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllRetryRetryableErr(nil))
	require.Nil(t, SetAllRetryRetryableErr(func(error) bool {
		return true
	}))
}

func TestSetRetryEmitter(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetRetryEmitter("retry3", metrics.Noop{})
	require.NotNil(t, err, "retry3 should not exist")

	err = SetRetryEmitter("retry2", nil)
	require.NotNil(t, err, "emitter should not be nil")

	err = SetRetryEmitter("retry2", metrics.Noop{})
	require.Nil(t, err)
}

func TestSetAllRetryEmitter(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllRetryEmitter(nil))
	require.Nil(t, SetAllRetryEmitter(metrics.Noop{}))
}

func TestWithDisabled(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetRetryRetryableErr("retry2", func(err error) bool { return true })
	require.Nil(t, err)

	ctx := trpc.BackgroundContext()
	msg := trpc.Message(ctx)

	msg.WithCalleeServiceName(serviceWelcome)
	msg.WithCalleeMethod(methodHello)

	require.False(t, disabled(ctx))

	var calledN int
	err = interceptor(ctx, nil, nil, func(ctx context.Context, req, rsp interface{}) error {
		calledN++
		return errors.New("need retry")
	})
	require.NotNil(t, err)
	require.Equal(t, 4, calledN)

	ctx = WithDisabled(ctx)
	require.True(t, disabled(ctx))

	calledN = 0
	err = interceptor(ctx, nil, nil, func(ctx context.Context, req, rsp interface{}) error {
		calledN++
		return errors.New("need retry")
	})
	require.NotNil(t, err)
	require.Equal(t, 1, calledN)
}

func TestDefaultCfg(t *testing.T) {
	cfg := `
service:
  - name: service.retry
    retry_hedging:
      retry:
        backoff:
          linear: [10ms, 15ms]
  - callee: service.hedging
    retry_hedging:
      hedging:
        hedging_delay: 10ms
`
	require.Nil(t, defaultManager.Setup("", yaml.NewDecoder(bytes.NewReader([]byte(cfg)))))
}

func patchDefaultManager() (func(), error) {
	f, err := os.Open(cfgFile)
	if err != nil {
		return nil, err
	}

	newDefaultManager := retryHedgingManager{}
	if err = newDefaultManager.Setup("", yaml.NewDecoder(f)); err != nil {
		return nil, err
	}

	oldDefaultManager := defaultManager
	defaultManager = &newDefaultManager

	return func() { defaultManager = oldDefaultManager }, nil
}

func TestSetRetryConditionalLog(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetRetryConditionalLog("retry3", &ConsoleLog{}, func(view.Stat) bool { return false }),
		"retry3 should not exist")
	require.Nil(t, SetRetryConditionalLog("retry2", &ConsoleLog{}, func(view.Stat) bool { return true }))

	require.NotNil(t, SetRetryConditionalCtxLog("retry2", nil, func(view.Stat) bool { return true }),
		"log cannot be nil")
	require.NotNil(t, SetRetryConditionalCtxLog("retry2", &CtxConsoleLog{}, nil),
		"condition cannot be nil")
	require.NotNil(t, SetRetryConditionalCtxLog("retry3", &CtxConsoleLog{}, func(view.Stat) bool {
		return true
	}), "retry3 should not exist")
	require.Nil(t, SetRetryConditionalCtxLog("retry2", &CtxConsoleLog{}, func(view.Stat) bool { return true }))
}

func TestSetAllRetryConditionalLog(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllRetryConditionalLog(nil, nil))
	require.NotNil(t, SetAllRetryConditionalLog(&ConsoleLog{}, nil))
	require.Nil(t, SetAllRetryConditionalLog(&ConsoleLog{}, func(stat view.Stat) bool {
		return true
	}))
	require.Nil(t, SetAllRetryConditionalCtxLog(&CtxConsoleLog{}, func(stat view.Stat) bool {
		return true
	}))
}

type ConsoleLog struct{}

func (ConsoleLog) Println(s string) {
	log.Println(s)
}

type CtxConsoleLog struct{}

func (CtxConsoleLog) Println(ctx context.Context, s string) {
	log.Println(s)
}

func TestSetRetryRspToErr(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	err = SetRetryRspToErr("retry3", func(rsp interface{}) error { return nil })
	require.NotNil(t, err, "retry3 should not exist")

	err = SetRetryRspToErr("retry2", func(rsp interface{}) error { return nil })
	require.Nil(t, err)
}

func TestSetAllRetryRspToErr(t *testing.T) {
	done, err := patchDefaultManager()
	require.Nil(t, err)
	defer done()

	require.NotNil(t, SetAllRetryRspToErr(nil))
	require.Nil(t, SetAllRetryRspToErr(func(rsp interface{}) error {
		return nil
	}))
}
