# mock

[![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/pcgtrpcproject/p-39d43cbb08d847a799ef26493deea97f/badge?X-DEVOPS-PROJECT-ID=pcgtrpcproject)](http://devops.oa.com:/ms/process/api-html/user/builds/projects/pcgtrpcproject/pipelines/p-39d43cbb08d847a799ef26493deea97f/latestFinished?X-DEVOPS-PROJECT-ID=pcgtrpcproject)[![Coverage](https://tcoverage.woa.com/api/getCoverage/getTotalImg/?pipeline_id=p-39d43cbb08d847a799ef26493deea97f)](http://macaron.oa.com/api/coverage/getTotalLink/?pipeline_id=p-39d43cbb08d847a799ef26493deea97f)[![GoDoc](https://img.shields.io/badge/API%20Docs-GoDoc-green)](http://godoc.oa.com/git.code.oa.com/trpc-go/trpc-filter/mock)
It implement back-end dependency interface mock calls via interceptors.

## Instructions for use

- Add import

```go
import (
   _ "git.code.oa.com/trpc-go/trpc-filter/mock"
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
