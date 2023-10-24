# transinfo blocker

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/transinfo-blocker.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/transinfo-blocker)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/transinfo-blocker)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/transinfo-blocker)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/transinfo-blocker.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/transinfo-blocker.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=transinfo-blocker&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/transinfo-blocker)


- trpc 框架下透传字段安全插件, 用于屏蔽调用下游的字段，避免登录态及其他敏感信息泄露问题。

## 使用说明

- `main.go` 增加 import

```go
import (
    _ "trpc.group/trpc-go/trpc-filter/transinfo-blocker"
)
```

- `trpc_go.yaml` 增加 client-filter, puigins 配置

```yaml
client:
  filter:
    - transinfo-blocker

plugins:
  security:
    transinfo-blocker:
      default: # 默认客户端调用配置，所有rpc调用未配置rpc_name_cfg会使用这个
        mode: blacklist # none, blocker, blacklist
        keys:
          - oidb_header
      rpc_name_cfg: # 单独命令字调用客户端配置, 会对于这个命令字覆盖default
        /trpc.qq_news.user_info.UserInfo/HandleProcess:
          mode: blocker
          keys: # mode=blocker, keys为空则所有都不透传
        /trpc.qq_news.user_info.UserInfo/Call:
          mode: blacklist
          keys:
            - trpc-trace
```
