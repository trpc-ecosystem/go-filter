# Validation

[![Go Reference](https://pkg.go.dev/badge/trpc.group/trpc-go/trpc-filter/validation.svg)](https://pkg.go.dev/trpc.group/trpc-go/trpc-filter/validation)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-filter/validation)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-filter/validation)
[![Tests](https://github.com/trpc-ecosystem/go-filter/actions/workflows/validation.yml/badge.svg)](https://github.com/trpc-ecosystem/go-filter/actions/workflows/validation.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-filter/branch/main/graph/badge.svg?flag=validation&precision=2)](https://app.codecov.io/gh/trpc-ecosystem/go-filter/tree/main/validation)

Parameter automatic validation plugin. Supports custom error codes.

## Usage

Import this plugin in your code.

```golang
import (
    _ "trpc.group/trpc-go/trpc-filter/validation"
)
```

Configure the trpc-go config file. In the server's filter configuration, enable the validation interceptor to automatically validate the req request parameters.

```yaml
server:
 ...
 filter:
  ...
  - validation
```

Configure the trpc-go config file. In the client's filter configuration, enable the validation interceptor to automatically validate the rsp response parameters.

```yaml
client:
 ...
 filter:
  ...
  - validation
```

Enable local interception log recording (optional).

```yaml
plugins:                     
  auth:
    validation:
      enable_error_log: true
```

Customize error codes (optional).

When req validation fails in the server filter, the default error code `errs.RetServerValidateFail 51` will be used.
When rsp validation fails in the client filter, the default error code `errs.RetClientValidateFail 151` will be used.

You can customize the error codes with the following configuration:

```yaml
plugins:
  auth:
    validation:
      enable_error_log: true
      server_validate_err_code: 100101
      client_validate_err_code: 100102
```

## Writing Proto Protocol Files

```protobuf
syntax = "proto3";

package trpc.test.helloworld;

import "validate/validate.proto";

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
