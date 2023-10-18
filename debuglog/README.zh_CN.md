# debuglog

启用此插件后，会对所有接口的请求进行日志打印。

## 使用说明

在代码中引入此插件。

```go
import (
   _ "trpc.group/trpc-go/trpc-filter/debuglog"
)
```

配置trpc-go框架配置。其中，`debuglog`可以替换为以下条目：
- debuglog：默认打印模式
- simpledebuglog：不打印包体
- pjsondebuglog：以格式化json打印包体
- jsondebuglog：以压缩型json打印包体

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

## 详细配置（可选）

```yaml
plugins:
  tracing:
    debuglog:
      log_type: simple # 默认日志打印方式
      err_log_level: error # 错误日志打印级别
      nil_log_level: info # 非错误日志打印级别
      server_log_type: prettyjson # server日志打印方式，会覆盖log_type的设定
      client_log_type: json # client日志打印方式，会覆盖log_type的设定
      enable_color: true # 开启日志显示不同颜色   默认false
      include: # 包含的匹配规则,当配置不为空时,exclude选项将无效.
        # method 和 retcode 选项同时指定时, 都命中才包含。
        - method: /trpc.app.server.service/methodA 
          retcode: 51
        # 只按方法名包含
        - method: /trpc.app.server.service/methodB
        # 只按错误码包含
        - retcode: 52
      exclude: # 忽略打印的匹配规则, 当 include 选项不为空时, 该项配置不生效。
        # method 和 retcode 选项同时指定时, 都命中才排除。
        - method: /trpc.app.server.service/methodC
          retcode: 53
        # 只按方法名排除
        - method: /trpc.app.server.service/methodD
        # 只按错误码排除
        - retcode: 54
```

- 请注意，使用插件配置时，filter项的配置必须为`debuglog`，否则插件的配置不会生效。
- log_type/server_log_type/client_log_type的可选项如下：
  - default：对应debuglog（默认）
  - simple：对应simpledebuglog，不打印包体
  - prettyjson：对应pjsondebuglog，以格式化json打印包体
  - json：对应jsondebuglog，以压缩型json打印包体
- err_log_level的可选项如下：
  - error（默认）
  - debug
  - info
  - warning
  - fatal
- nil_log_level的可选性如下：
  - debug（默认）
  - info
  - warning
  - error
  - fatal

## 自定义打印方法（可选）

- 如果用户需要自定义请求回包的打印方法，可以通过自行注册自定义的打印方法来实现。
- 请注意，使用自定义打印方式时，filter必须为`debuglog`，并会覆盖插件配置中的配置。

```go
import (
	"context"
	"fmt"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-filter/debuglog"
)

func main() {
	// 自定义Server打印函数
	debugServerLogFunc := func(ctx context.Context, req, rsp interface{}) string {
		return fmt.Sprintf(", req:%+v, rsp:%+v, this is server log test", req, rsp)
	}
	// 自定义Client打印函数
	debugClientLogFunc := func(ctx context.Context, req, rsp interface{}) string {
		return fmt.Sprintf(", req:%+v, rsp:%+v, this is client log test", req, rsp)
	}
	// 注册filter
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
