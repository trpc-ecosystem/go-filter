# trpc-go tvar统计插件

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/tvar.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/tvar)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/tvar)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/tvar)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/tvar.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/tvar.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=tvar&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/tvar)

## 插件简介

实现serverside、clientside RPC等监控项的统计上报

注册admin接口：/cmds/stats/rpc 接口可查询监控项，及支持情况

| 监控项                                 | 必须/可选 | 说明                      | 已支持 |
|:------------------------------------|:------|:------------------------|:----|
| rpc_revision                        | 必须    | 版本                      | ❌ | 
| rpc_frame_thread_count              | 非必须   | rpc框架开辟的线程数             | ✔ |
| rpc_service_count                   | 非必须   | 当前服务有几个service          | ✔ |
| rpc_service_xxx_connection_count    | 必须    | service连接数              | ✔ |
| rpc_service_xxx_req_total           | 必须    | service累积收到的总请求数        | ✔ |
| rpc_service_xxx_req_active          | 必须    | service正在处理的请求数         | ✔ |
| rpc_service_xxx_rsp_total           | 必须    | service累积返回的回包数         | ✔ |
| rpc_service_xxx_req_avg_len         | 必须    | service收到请求包的平均大小       | ❌ |
| rpc_service_xxx_rsp_avg_len         | 必须    | service回包的平均大小          | ❌ |
| rpc_service_xxx_error_total         | 必须    | sevice错误数               | ✔ |
| rpc_service_xxx_business_error      | 非必须   | service业务代码返回错误数        | ✔ |
| rpc_service_xxx_protocol_error      | 非必须   | sevice协议错误              | ❌ |
| rpc_service_xxx_latency_p1          | 必须    | sevice 百分比延时，用户可以自定义分位值 | ✔ |
| rpc_service_xxx_latency_p2          | 必须    | sevice 百分比延时，用户可以自定义分位值 | ✔ |
| rpc_service_xxx_latency_p3          | 必须    | sevice 百分比延时，用户可以自定义分位值 | ✔ |
| rpc_service_xxx_latency_999         | 必须    | sevice p999延时           | ✔ |
| rpc_service_xxx_latency_9999        | 必须    | sevice p9999延时          | ✔ |
| rpc_service_xxx_latency_avg/max/min | 必须    | sevice 延时的avg max min   | ❌ |
| rpc_service_xxx_qps                 | 必须    | sevice qps              | ✔ |
| rpc_client_xxx_connection_count     | 必须    | client有几个连接             | ✔  |
| rpc_client_xxx_req_total            | 必须    | client累积发出的请求数          | ✔ |
| rpc_client_xxx_req_active           | 必须    | client等待回复的请求数          | ✔ |
| rpc_client_xxx_rsp_total            | 必须    | client累积收到的回复数          | ✔ |
| rpc_client_xxx_error_total          | 必须    | client收到的错误结果数          | ✔ |
| rpc_client_xxx_latency_p1           | 必须    | client 百分位延时，用户可以自定义分位值 | ✔ | 
| rpc_client_xxx_latency_p2           | 必须    | client 百分位延时，用户可以自定义分位值 | ✔ |
| rpc_client_xxx_latency_p3           | 必须    | client 百分位延时，用户可以自定义分位值 | ✔ |
| rpc_client_xxx_latency_99           | 必须    | client p99延时            | ✔ | 
| rpc_client_xxx_latency_999          | 必须    | client p999延时           | ✔ |
| rpc_client_xxx_latency_avg/max/min | 必须    | client延时的avg/max/min    | ❌ |

## 使用方式

1. 代码中导入包：`import _ "trpc.group/trpc-go/trpc-filter/tvar"
2. trpc_go.yaml中增加初始化配置

   ```yaml
   plugins:
     apm:
       tvar:
         percentile:
           - p50
           - p90
           - p99
   ```

3. 查询：`curl http://ip:adminport/cmds/stats/rpc`
