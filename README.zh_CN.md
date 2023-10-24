[English](README.md) | 中文

# go-filter

[![LICENSE](https://img.shields.io/badge/license-Apache--2.0-green.svg)](https://github.com/trpc-ecosystem/go-filter/blob/main/LICENSE)

本仓库提供了一些常用的 trpc-go 拦截器，如：

* debuglog: 自动打印客户端/服务端接口的请求和响应
* degrade: 服务端熔断限流器
* filterextensions: 拦截器功能扩展，支持到 method 粒度
* hystrix: 基于 Netflix 开源的 hystrix 实现的服务端熔断限流器
* jwt: 用户身份验证拦截器
* masking: 敏感数据脱敏模块
* mock: 故障模拟
* recovery: 服务端 panic 自动捕获插件
* referer: web referer 验证
* slime: 重试/对冲请求插件
* transinfo-blocker: 透传参数安全插件，避免敏感信息泄露
* tvar: 监控项统计上报
* validation: 参数自动校验插件
