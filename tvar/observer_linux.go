// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

//go:build linux
// +build linux

package tvar

import (
	"context"
	"os"

	"github.com/bastjan/netstat"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/log"
)

func asyncUpdateTCPConnection() {
	cfg := trpc.GlobalConfig()
	ports := map[uint16]bool{}
	for _, s := range cfg.Server.Service {
		ports[s.Port] = true
	}

	conns, err := netstat.TCP.Connections()
	if err != nil {
		log.Errorf("list tcp connections fail: %v", err)
		return
	}

	pid := os.Getpid()

	for _, conn := range conns {
		if conn.Pid != pid {
			continue
		}
		_, ok := ports[uint16(conn.Port)]
		if !ok { // client发起到其他服务的tcpconn
			clientConnNum.Observe(context.Background(), 1)
		} else { // 可以粗略认为是接受client请求建立的tcpconn
			serviceConnectionNum.Observe(context.Background(), 1)
		}
	}
}
