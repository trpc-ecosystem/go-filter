# recovery

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/recovery.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/recovery)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/recovery)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/recovery)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/recovery.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/recovery.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=recovery&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/recovery)

启用本插件后，会在请求入口处捕获panic并返回500，避免因为panic导致的程序崩溃。

一般情况下，本插件应配置为filter配置中的第一项。

## 使用说明

在代码中引入本插件。

```golang
import (
   _ "trpc.group/trpc-go/trpc-filter/recovery"
)
```

配置trpc-go框架配置文件。在server的filter配置中，按如下方法开启recovery拦截器。

```yaml
server:
 ...
 filter:
  - recovery 
  ...
```
