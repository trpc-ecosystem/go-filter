# Masking
tRPC敏感数据脱敏模块



## 使用说明

 - 增加import

````
import (
   _ "trpc.group/trpc-go/trpc-filter/masking"
)
````

 - TRPC框架配置文件，**client开启validation拦截器，自动对rsp响应参数值做脱敏**

````
client:
 ...
 filter:
  ...
  - masking 
````

 - TRPC框架配置文件，server开启validation拦截器，自动对req输入参数值做脱敏

````
server:
 ...
 filter:
  ...
  - masking 
````



## 编写proto协议文件

```
syntax = "proto3";

package trpc.test.helloworld;

import "masking.proto";

option go_package="trpc.group/trpcprotocol/test/helloworld";

... // 省略部分proto消息结构

message SearchReply {
  string query = 1;
  string phone_num= 2 [(masking.rules).string.mobile = true];
}
```