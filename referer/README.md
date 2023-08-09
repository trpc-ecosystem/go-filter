# referer

[![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/pcgtrpcproject/p-a437407d143f42f6b408ec8f23874fd5/badge?X-DEVOPS-PROJECT-ID=pcgtrpcproject)](http://devops.oa.com:/ms/process/api-html/user/builds/projects/pcgtrpcproject/pipelines/p-a437407d143f42f6b408ec8f23874fd5/latestFinished?X-DEVOPS-PROJECT-ID=pcgtrpcproject) [![Coverage](https://tcoverage.woa.com/api/getCoverage/getTotalImg/?pipeline_id=p-a437407d143f42f6b408ec8f23874fd5)](http://macaron.oa.com/api/coverage/getTotalLink/?pipeline_id=p-a437407d143f42f6b408ec8f23874fd5) [![GoDoc](https://img.shields.io/badge/API%20Docs-GoDoc-green)](http://godoc.oa.com/git.code.oa.com/trpc-go/trpc-filter/referer)

http referer 安全验证

## 使用说明

- 增加 import

```golang
import (
   _ "git.code.oa.com/trpc-go/trpc-filter/referer"
)
```

```yaml
server:
 ...
 filter:
  ...
  - referer
plugins:
  auth:
    referer:
      apply_to_all_path:
        - qq.com
      path1:
        - qq.com
        - oa.com
      path2:
        - NULL

```

## 配置说明

- apply_to_all_path 默认规则
- NULL 表示 referer 为空时可以放过
- `*` 匹配所有的(不安全)
- referer 插件 会对所有 url 验证 非白名单模式
