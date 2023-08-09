# recovery

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
