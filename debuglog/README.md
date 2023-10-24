# debuglog

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/debuglog.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/debuglog)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/debuglog)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/debuglog)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/debuglog.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/debuglog.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=debuglog&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/debuglog)

Enable this plugin to log the requests of all API.

## Usage

Import this plugin in your code.

```go
import (
   _ "trpc.group/trpc-go/trpc-filter/debuglog"
)
```

Configure the trpc-go framework. Replace `debuglog` with one of the following items:

- debuglog: Default logging mode.
- simpledebuglog: Do not log the request body.
- pjsondebuglog: Log the request body in formatted JSON.
- jsondebuglog: Log the request body in compressed JSON.

```yaml
server:
 ...
 filter:
  ...
  - debuglog

client:
 ...
 filter:
  ...
  - debuglog
```

## Detailed Configuration (Optional)

```yaml
plugins:
  tracing:
    debuglog:
      log_type: simple # Default log printing method
      err_log_level: error # Error log printing level
      nil_log_level: info # Non-error log printing level
      server_log_type: prettyjson # Server log printing method, overrides log_type setting
      client_log_type: json # Client log printing method, overrides log_type setting
      enable_color: true # Enable displaying logs in different colors. Default is false.
      include: # Matching rules to include. When configured, exclude option will be ignored.
        # Include when both method and retcode match.
        - method: /trpc.app.server.service/methodA
          retcode: 51
        # Include by method name only.
        - method: /trpc.app.server.service/methodB
        # Include by error code only.
        - retcode: 52
      exclude: # Matching rules to exclude logs. This configuration is ignored when include option is not empty.
        # Exclude when both method and retcode match.
        - method: /trpc.app.server.service/methodC
          retcode: 53
        # Exclude by method name only.
        - method: /trpc.app.server.service/methodD
        # Exclude by error code only.
        - retcode: 54
```

- Note that when using plugin configuration, the filter option must be set to debuglog, otherwise the plugin configuration will not take effect.
- The options for log_type, server_log_type, and client_log_type are as follows:
    - default: Corresponds to debuglog (default)
    - simple: Corresponds to simpledebuglog, does not log the request body
    - prettyjson: Corresponds to pjsondebuglog, logs the request body in formatted JSON
    - json: Corresponds to jsondebuglog, logs the request body in compressed JSON
- The options for err_log_level are as follows:
    - error (default)
    - debug
    - info
    - warning
    - fatal
- The options for nil_log_level are as follows:
    - debug (default)
    - info
    - warning
    - error
    - fatal

## Custom Print Methods (Optional)

- If you need to customize the print method for request and response, you can register your own print method.
- Note that when using a custom print method, the filter must be set to debuglog, and it will override the configuration in the plugin.

```go
import (
	"context"
	"fmt"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-filter/debuglog"
)

func main() {
    // Custom server log function
	debugServerLogFunc := func(ctx context.Context, req, rsp interface{}) string {
		return fmt.Sprintf(", req:%+v, rsp:%+v, this is server log test", req, rsp)
	}
    // Custom client log function
	debugClientLogFunc := func(ctx context.Context, req, rsp interface{}) string {
		return fmt.Sprintf(", req:%+v, rsp:%+v, this is client log test", req, rsp)
	}
    // Register the filter
	filter.Register("debuglog",
		debuglog.ServerFilter(debuglog.WithLogFunc(debugServerLogFunc)),
		debuglog.ClientFilter(debuglog.WithLogFunc(debugClientLogFunc)))

	s := trpc.NewServer()

	pb.RegisterHttp_helloworldService(s, &Http_helloworldServerImpl{})
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}

}
```
