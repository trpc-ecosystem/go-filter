# tRPC-Go [jwt] 用户身份验证拦截器

## jwt 插件使用说明

### I. 必要条件

- 匿名导入该插件：

```go
import (
_ "trpc.group/trpc-go/trpc-filter/jwt"
)
```

- TRPC框架配置文件, 开启 jwt 拦截器

```yaml
server:
  ...
  filter:
    ...
    - jwt
plugins:
  auth:
    jwt:
      secret: q7wt3n1t # 私钥
      expired: 3600    # jwt签名过期时间，单位秒
      issuer: tencent  # 签发者
      exclude_paths:   # [可选] path白名单-如登陆接口不进行jwt校验
        - /v1/login
```

### II. 可选

- 校验成功后，通过 `GetCustomInfo`方法，可以从 ctx 中获取用户信息

```go
// GetCtxAuthInfo 从 ctx 中获取用户信息
func GetCtxAuthInfo(ctx context.Context) *AuthInfo {
    var authInfo = &AuthInfo{}
    if err := jwt.GetCustomInfo(ctx, authInfo); err != nil {
        return err
    }
    return authInfo
}
```

- 用户可以自定义 token 的逻辑，覆盖 `jwt.DefaultParseTokenFunc` 函数即可。

```go
jwt.DefaultParseTokenFunc = func(ctx context.Context, req interface{}) (string, error) {
    head := trpcHttp.Head(ctx)
    token := head.Request.Header.Get("Authorization")
    return strings.TrimPrefix(token, "Bearer "), nil
}
```

- 更进一步，用户可以通过实现 `Signer` 接口，来实现自己的签名生成、验证逻辑； 然后通过 `SetDefaultSigner(s)`方法覆盖默认的签名器。
```go
// Signer 数字签名器接口
type Signer interface {
	Sign(custom interface{}) (string, error)
	Verify(token string) (interface{}, error)
}

SetDefaultSigner(NewCustomSigner())
```