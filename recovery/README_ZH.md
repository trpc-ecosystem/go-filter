# recovery

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
