# recovery

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/recovery.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/recovery)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/recovery)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/recovery)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/recovery.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/recovery.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=recovery&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/recovery)

Enable this plugin to capture panics at the entry point of requests and return a 500 response, preventing program crashes caused by panics.

In general, this plugin should be configured as the first item in the filter configuration.

## Usage

Import this plugin in your code.

```golang
import (
   _ "trpc.group/trpc-go/trpc-filter/recovery"
)
```

Configure the trpc-go framework. In the server's filter configuration, enable the recovery interceptor as follows:

```yaml
server:
  ...
  filter:
    - recovery
    ...
```
