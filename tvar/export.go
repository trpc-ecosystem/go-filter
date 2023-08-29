// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package tvar

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"trpc.group/trpc-go/trpc-go/admin"
	"trpc.group/trpc-go/trpc-go/log"

	"trpc.group/trpc-go/trpc-filter/tvar/meterprovider"
)

func start() {
	admin.HandleFunc("/cmds/stats/rpc", handler)
	go ticker(time.NewTicker(time.Second * 5))
}

func handler(w http.ResponseWriter, _ *http.Request) {
	all := dumpRPCMetrics()
	b, err := json.Marshal(all)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(b)
}

func ticker(ticker *time.Ticker) {
	for range ticker.C {
		// 更新tcp连接数统计
		asyncUpdateTCPConnection()

		// 更新服务qps统计
		updateServiceQPS()
	}
}

// 我们采用累计的统计方式，非delta的统计方式，所以这里记录下上次exporter导出的数据，
// 如果接下来一段时间没有rpc，exporter将收集不到新数据，我们将使用这里的数据代替
var snapshot []meterprovider.ExportRecord

func updateServiceQPS() {
	qps := int64(float64(serviceQWin.Count()) / float64(serviceQWin.Size().Seconds()))
	serviceQPS.Observe(context.Background(), qps)
}

func dumpRPCMetrics() []string {
	// collect first
	if err := exporter.Collect(context.Background()); err != nil {
		log.Errorf("collect fail: %v", err)
	}

	// then we got records
	records := exporter.GetRecords()
	if len(records) != 0 {
		snapshot = records
	}

	// prepare to export
	var all []string
	for _, r := range snapshot {
		if r.AggregationKind != aggregation.HistogramKind {
			all = append(all, r.String())
			continue
		}

		var msg strings.Builder
		msg.WriteString(r.String())

		lh := latencyHistogram{r.Histogram}

		for _, percentile := range tvarCfg.Percentile {
			// convert p99 to 0.99
			p := strings.ToLower(percentile)
			v := strings.TrimPrefix(p, "p")
			n, err := strconv.Atoi(v)
			if err != nil {
				panic("invalid percentile value, should be format: pNN")
			}
			fv := float64(n) / math.Pow10(len(v))
			// then calculate the percentile
			pv := lh.Percentile(fv)
			msg.WriteString(fmt.Sprintf("%s:%.2f ", percentile, pv))
		}
		all = append(all, msg.String())
	}
	return all
}
