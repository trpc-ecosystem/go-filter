// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package debuglog

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/plugin"
)

const configInfo = `
plugins:
  tracing:
    debuglog:
      log_type: simple
      err_log_level: error
      nil_log_level: info
      server_log_type: prettyjson
      client_log_type: json
      enable_color: true
      exclude:
        - method: /trpc.app.server.service/method
        - retcode: 51
`

type testReq struct {
	A int
	B string
}

type testRsp struct {
	C int
	D string
}

// TestFilter_PluginType is the unit test for the Log interceptor plugin type.
func TestPlugin_PluginType(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, pluginType, p.Type())
}

// TestPlugin_Setup is the unit test for setting the attributes of the Log interceptor plugin.
func TestPlugin_Setup(t *testing.T) {
	cfg := trpc.Config{}
	err := yaml.Unmarshal([]byte(configInfo), &cfg)
	assert.Nil(t, err)

	conf, ok := cfg.Plugins[pluginType][pluginName]
	if !ok {
		assert.Nil(t, conf)
	}

	p := &Plugin{}
	err = p.Setup(pluginName, &plugin.YamlNodeDecoder{Node: &conf})
	assert.Nil(t, err)
}

func TestFilter_Filter(t *testing.T) {
	rsp := testRsp{
		C: 456,
		D: "456",
	}
	testHandleFunc1 := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &rsp, nil
	}
	testHandleFunc2 := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errs.New(errs.RetServerSystemErr, "system error")
	}

	testHandleFunc3 := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errs.New(errs.RetServerSystemErr, "system error")
	}

	testHandleFunc4 := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errs.New(errs.RetServerDecodeFail, "system decode error")
	}

	ctx := trpc.BackgroundContext()
	msg := trpc.Message(ctx)
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:6379")
	msg.WithRemoteAddr(addr)
	msg.WithServerRPCName("/trpc.app.server.service/methodA")
	req := testReq{
		A: 123,
		B: "123",
	}

	ex := newTestRule("/trpc.app.server.service/methodA", errs.RetClientCanceled)
	in := newTestRule("/trpc.app.server.service/methodA", errs.RetServerSystemErr)
	sf := ServerFilter(
		WithExclude(ex), WithInclude(in),
		WithInclude(newTestRule("/trpc.app.server.service/methodA", errs.RetOK)))

	_, err := sf(ctx, req, testHandleFunc1)
	assert.NotNil(t, sf)
	assert.Nil(t, err)

	ret, err := ServerFilter(WithExclude(ex))(ctx, req, testHandleFunc1)
	assert.NotNil(t, ret)
	assert.Nil(t, err)

	_, err = sf(ctx, req, testHandleFunc2)
	assert.NotNil(t, err)

	deadLineCtx, deadLineCancel := context.WithDeadline(ctx, time.Now().Add(time.Second*1))
	defer deadLineCancel()
	_, err = sf(deadLineCtx, req, testHandleFunc3)
	assert.NotNil(t, err)

	_, err = sf(deadLineCtx, req, testHandleFunc4)
	assert.NotNil(t, err)

	testClientHandleFunc1 := func(ctx context.Context, req, rsp interface{}) error {
		return nil
	}
	testClientHandleFunc2 := func(ctx context.Context, req, rsp interface{}) error {
		return errs.New(errs.RetClientConnectFail, "connect fail")
	}
	cf := ClientFilter()
	assert.NotNil(t, cf)
	assert.Nil(t, cf(ctx, req, rsp, testClientHandleFunc1))
	assert.NotNil(t, cf(ctx, req, rsp, testClientHandleFunc2))
}

func TestFilter_LogFunc(t *testing.T) {
	req := testReq{
		A: 123,
		B: "123",
	}

	rsp := testRsp{
		C: 456,
		D: "456",
	}

	ctx := trpc.BackgroundContext()

	assert.Equal(
		t, DefaultLogFunc(ctx, req, rsp),
		", req:{A:123 B:123}, rsp:{C:456 D:456}",
	)

	assert.Equal(t, SimpleLogFunc(ctx, req, rsp), "")

	assert.Equal(
		t, PrettyJSONLogFunc(ctx, req, rsp),
		"\nreq:{\n  \"A\": 123,\n  \"B\": \"123\"\n}\nrsp:{\n  \"C\": 456,\n  \"D\": \"456\"\n}",
	)

	assert.Equal(
		t, JSONLogFunc(ctx, req, rsp),
		"\nreq:{\"A\":123,\"B\":\"123\"}\nrsp:{\"C\":456,\"D\":\"456\"}",
	)
}

func Test_getLogFunc(t *testing.T) {
	got := getLogFunc("simple")
	gv := reflect.ValueOf(got)
	wv := reflect.ValueOf(SimpleLogFunc)
	assert.Equal(t, gv.Pointer(), wv.Pointer())

	got = getLogFunc("json")
	gv = reflect.ValueOf(got)
	wv = reflect.ValueOf(JSONLogFunc)
	assert.Equal(t, gv.Pointer(), wv.Pointer())

	got = getLogFunc("prettyjson")
	gv = reflect.ValueOf(got)
	wv = reflect.ValueOf(PrettyJSONLogFunc)
	assert.Equal(t, gv.Pointer(), wv.Pointer())

	got = getLogFunc("default")
	gv = reflect.ValueOf(got)
	wv = reflect.ValueOf(DefaultLogFunc)
	assert.Equal(t, gv.Pointer(), wv.Pointer())
}

func Test_getLogLevelFunc(t *testing.T) {
	got := getLogLevelFunc("info", "debug")
	gv := reflect.ValueOf(got)
	wv := reflect.ValueOf(LogContextfFuncs["info"])
	assert.Equal(t, gv.Pointer(), wv.Pointer())

	got = getLogLevelFunc("debug", "debug")
	gv = reflect.ValueOf(got)
	wv = reflect.ValueOf(LogContextfFuncs["debug"])
	assert.Equal(t, gv.Pointer(), wv.Pointer())

	got = getLogLevelFunc("trace", "debug")
	gv = reflect.ValueOf(got)
	wv = reflect.ValueOf(LogContextfFuncs["trace"])
	assert.Equal(t, gv.Pointer(), wv.Pointer())

	got = getLogLevelFunc("default", "debug")
	gv = reflect.ValueOf(got)
	wv = reflect.ValueOf(LogContextfFuncs["debug"])
	assert.Equal(t, gv.Pointer(), wv.Pointer())
}

func Test_options_passed(t *testing.T) {
	type fields struct {
		include []*RuleItem
		exclude []*RuleItem
	}
	type args struct {
		rpcName string
		errCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "include rule empty - passed",
			fields: fields{
				exclude: []*RuleItem{newTestRule("ext-method-1", 123)},
			},
			args: args{
				rpcName: "method-1",
				errCode: 123,
			},
			want: true,
		},
		{
			name: "include rule empty - not passed",
			fields: fields{
				exclude: []*RuleItem{newTestRule("ext-method-1", 123)},
			},
			args: args{
				rpcName: "ext-method-1",
				errCode: 123,
			},
			want: false,
		},
		{
			name: "has include rule item - passed",
			fields: fields{
				include: []*RuleItem{newTestRule("in-method-1", 123)},
			},
			args: args{
				rpcName: "in-method-1",
				errCode: 123,
			},
			want: true,
		},
		{
			name: "has include rule item - not passed",
			fields: fields{
				include: []*RuleItem{newTestRule("in-method-1", 123)},
			},
			args: args{
				rpcName: "in-method-1",
				errCode: 0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options{
				include: tt.fields.include,
				exclude: tt.fields.exclude,
			}
			if got := o.passed(tt.args.rpcName, tt.args.errCode); got != tt.want {
				t.Errorf("options.passed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newTestRule(method string, retcode int) *RuleItem {
	return &RuleItem{Method: &method, Retcode: &retcode}
}

func Test_getLogFormat(t *testing.T) {
	logFormat := getLogFormat(debugLevel, false, "")
	assert.Equal(t, logFormat, "")
	logFormat = getLogFormat(debugLevel, true, "server request:%s, cost:%s, from:%s%s")
	assert.Equal(t, logFormat, "\033[1;35mserver request:%s, cost:%s, from:%s%s\033[0m")
}

func Test_fontColor(t *testing.T) {
	var blue color = 34
	assert.Equal(t, blue.fontColor(), uint8(34))
}

func Test_WithEnableColor(t *testing.T) {
	WithEnableColor(false)
}
