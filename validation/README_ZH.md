# Validation

参数自动校验插件，支持自定义错误码。

## 使用说明

在代码中引入本插件。

```golang
import (
   _ "trpc.group/trpc-go/trpc-filter/validation"
)
```

配置trpc-go框架配置文件。在server的filter配置中，按如下方法开启validation拦截器，自动校验req请求参数。

```yaml
server:
 ...
 filter:
  ...
  - validation
```

配置trpc-go框架配置文件。在client的filter配置中，按如下方法开启validation拦截器，自动校验rsp请求参数。

```yaml
client:
 ...
 filter:
  ...
  - validation
```

开启本地拦截日志记录（可选）

```yaml
plugins:                     
  auth:
    validation:
      enable_error_log: true # 开启本地日志记录
```

自定义错误码 （可选）

- 在 server filter 校验 req 失败时，默认将使用错误码 errs.RetServerValidateFail 51。
- 在 client filter 校验 rsp 失败时，默认将使用错误码 errs.RetClientValidateFail 151。

可以通过如下配置对错误码进行自定义：

```yaml
plugins:
  auth:
    validation:
      enable_error_log: true
      server_validate_err_code: 100101
      client_validate_err_code: 100102
```

## 编写proto协议文件

更加详细的指引请参考：<https://git.woa.com/devsec/protoc-gen-secv>

```protobuf
syntax = "proto3";

package trpc.test.helloworld;

import "trpc/common/validate.proto";

option go_package="trpc.group/trpcprotocol/test/helloworld";

/* SearchRequest represents a search query, with pagination options to
 * indicate which results to include in the response.
 * Hint use https://regex-golang.appspot.com/assets/html/index.html for
 *  Regex validation in Go
 */

message SearchRequest {
  string query = 1 [(validate.rules).string = {
                      pattern:   "([A-Za-z]+) ([A-Za-z]+)*$",
                      max_bytes: 50,
                   }];
  string email_1= 2 [(validate.rules).string.alphabets = true];
  string email_2= 3 [(validate.rules).string.alphanums = true];
  string email_3= 4 [(validate.rules).string.lowercase = true];
}
```
