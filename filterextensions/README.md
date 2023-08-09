[![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/pcgtrpcproject/p-7c6d21257da948469e47d9ed3b4845ff/badge?X-DEVOPS-PROJECT-ID=pcgtrpcproject)](http://devops.oa.com/process/api-html/user/builds/projects/pcgtrpcproject/pipelines/p-7c6d21257da948469e47d9ed3b4845ff/latestFinished?X-DEVOPS-PROJECT-ID=pcgtrpcproject)
[![Coverage](https://tcoverage.woa.com/api/getCoverage/getTotalImg/?pipeline_id=p-7c6d21257da948469e47d9ed3b4845ff)](http://macaron.oa.com/api/coverage/getTotalLink/?pipeline_id=p-7c6d21257da948469e47d9ed3b4845ff)
[![GoDoc](https://img.shields.io/badge/API%20Docs-GoDoc-green)](http://godoc.oa.com/git.code.oa.com/trpc-go/trpc-filter/filter_extensions)

## tRPC-Go 拦截器扩展

tRPC-Go 支持在 `trpc_go.yaml` 中配置拦截器，但是拦截器的精度只到 service 层，无法作更细粒度的配置，如为 method 配置拦截器。

这个插件扩展了 `trpc_go.yaml` 的解析方式，以「插件」+「拦截器」的方式实现了 method 粒度的拦截器。

### 使用方式

匿名导入该插件：
```go
import _ "git.code.oa.com/trpc-go/trpc-filter/filterextensions"
```

`trpc_go.yaml` 中增加以下配置：
```yaml
server:
  # method_filters 配在 server 的全局拦截器中，最终拦载器的执行顺序为
  # method_filters
  #   server_method_1_filter_a
  #     server_method_1_filter_b
  #       global_server_filter
  #         filter_for_server_service_1
  filter: [method_filters, global_server_filter]
  service: &server_service # 这是 yaml 的引用语法
    - name: server_service_1
      filter: [filter_for_server_service_1]
      methods: # 这是 server 配置中新加的 methods 选项
        - name: method_1 # 方法名
          filters: # method 的 filter 列表
            - server_method_1_filter_a # 这些 filter 必须提前调用主库的 filter.Register 来注册
            - server_method_1_filter_b
client:
  # method_filters 配在 client_service_1 的局部拦截器中，最终拦截器的执行顺序为
  # global_client_filter
  #   filter_for_client_service_1
  #     method_filters
  #       client_method_1_filter_a
  #         client_method_1_filter_b
  filter: [global_client_filter]
  service: &client_service # 这是 yaml 的引用语法
    - name: client_service_1
      filter: [filter_for_client_service_1, method_filters]
      methods: # 这是 client 配置中新加的 methods 选项
        - name: method_1 # 方法名
          filters: # method 的 filter 列表
            - client_method_1_filter_a # 这些 filter 必须提前调用主库的 filter.Register 来注册
            - client_method_1_filter_b

plugins:
  filter_extensions: # 必填
    method_filters: # 必填
      client: *client_service # 选填，这里引用了 client.service 的配置。
      server: *server_service # 选填，这里引用了 server.service 的配置。
```