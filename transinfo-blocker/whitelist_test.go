// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// @Title  whitelist_test.go
// @Description  测试client-filter生效情况
// @Author  radaren 2020.06.02
// @Update  radaren 2020.06.02
package blocker

import (
	"context"
	"errors"
	"testing"

	yaml "gopkg.in/yaml.v3"

	"git.code.oa.com/trpc-go/trpc-go"
	"git.code.oa.com/trpc-go/trpc-go/plugin"
)

var (
	errRet      = errors.New("")
	metadataKey = "trpc-trace"
)

func handler(ctx context.Context, req interface{}, rsp interface{}) (err error) {
	return errRet
}

func checkK(ctx context.Context, req interface{}, rsp interface{}) (err error) {
	msg := trpc.Message(ctx)
	metadata := msg.ClientMetaData()
	if _, ok := metadata[metadataKey]; ok {
		return nil
	}
	return errRet
}

func TestFilter(t *testing.T) {
	ctx := trpc.BackgroundContext()
	err := ClientFilter(ctx, nil, nil, handler)
	if err != errRet {
		t.Errorf("unmatch Error:%+v, %+v", err, errRet)
	}

	msg := trpc.Message(ctx)
	metaData := make(map[string][]byte)
	metaData[metadataKey] = []byte{}
	metaData[metadataKey+"_"] = []byte{}
	msg.WithClientMetaData(metaData)

	cfg = &Config{}
	err = ClientFilter(ctx, nil, nil, handler)
	if err != errRet {
		t.Errorf("unmatch Error:%+v, %+v", err, errRet)
	}

	node := &struct {
		TransinfoBlocker yaml.Node `yaml:"transinfo-blocker"`
	}{}

	err = yaml.Unmarshal([]byte(testyaml), node)
	if err != nil {
		t.Error("unmarshal yaml failed:" + err.Error())
	}
	decoder := &plugin.YamlNodeDecoder{Node: &node.TransinfoBlocker}
	err = (&TransinfoBlocker{}).Setup("", decoder)
	if err != nil {
		t.Error("setup failed:" + err.Error())
	}
	msg.WithClientRPCName("/trpc.qq_news.user_info.UserInfo/Call")
	err = ClientFilter(ctx, nil, nil, checkK)
	if err != errRet {
		t.Errorf("unmatch Error:%+v, %+v", err, errRet)
	}
	msg.WithClientRPCName("/trpc.qq_news.user_info.UserInfo/HandleProcess")
	err = ClientFilter(ctx, nil, nil, checkK)
	if err != errRet {
		t.Errorf("unmatch Error:%+v, %+v", err, errRet)
	}
}
