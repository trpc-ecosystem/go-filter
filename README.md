English | [中文](README.zh_CN.md)

# go-filter

This repository provides several commonly used trpc-go filters, including:

* debuglog: Automatically logs the requests and responses of client/server interfaces.
* degrade: Server-side circuit breaker and rate limiter.
* filterextensions: Interceptor function extensions that support granularity down to the method level.
* hystrix: Server-side circuit breaker and rate limiter based on the open-source hystrix library by Netflix.
* jwt: User authentication interceptor.
* masking: Sensitive data masking module.
* mock: Fault simulation.
* recovery: Server-side panic automatic recovery plugin.
* referer: Web referer validation.
* slime: Retry/compensation request plugin.
* transinfo-blocker: Transparent parameter security plugin to prevent sensitive information leakage.
* tvar: Monitoring item statistics reporting.
* validation: Automatic parameter validation plugin.
