# mock

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/mock.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/mock)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/mock)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/mock)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/mock.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/mock.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=mock&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/mock)

It implement back-end dependency interface mock calls via interceptors.

## Instructions for use

- Add import

```go
import (
   _ "trpc.group/trpc-go/trpc-filter/mock"
)
```

- TRPC framework configuration file

```yaml
client:
 ...
 filter:
  ...
  - mock

plugins:
  tracing:
    mock:
      - method: /trpc.app.server.service/method   # mock the specified interface, or mock all interfaces if none is specified
        percent: 20   # Triggered by 20% chance
        delay: 10  # Delay 10ms
        timeout: true # Simulate timeout failure
        retcode: 111  # Simulation returns specific error codes
        retmsg: "error msg" # Simulation returns specific error messages
        body: '{"key":"value"}' # The simulation returns specific packet data, text type can be represented by json, binary data needs to be base64 encoded first
        serialization: 2 # Serialization method used for package data pb:0 jce:1 json:2, pb is used by default
```
