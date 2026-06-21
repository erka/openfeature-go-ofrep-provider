# OpenFeature Remote Evaluation Protocol Provider

This is a fork of `open-feature/go-sdk-contrib/pkg/providers/ofrep@0.1.5` and the **opinionated** Go implementation of the OFREP provider.

The package contains two providers, which can be used to interact with the OFREP single and bulk flag evaluation endpoint. The goal is to support the drop-in replacement of the existing OFREP GO provider why waiting for official implementation.

## Usage
Add a line to your `go.mod` file

```
replace github.com/open-feature/go-sdk-contrib/providers/ofrep => github.com/erka/openfeature-go-ofrep-provider v0.0.1
```

## Configuration

You can configure the provider using following configuration options,

| Configuration option | Details                                                                                                                 |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| WithApiKeyAuth       | Set the token to be used with "X-API-Key" header                                                                        |
| WithBearerToken      | Set the token to be used with "Bearer" HTTP Authorization schema                                                        |
| WithClient           | Provide a custom, pre-configured http.Client for OFREP service communication                                            |
| WithHeaderProvider   | Register a custom header provider for OFREP calls. You may utilize this for custom authentication/authorization headers |
| WithHeader           | Set a custom header to be used for authorization                                                                        |
| WithBaseURI          | Set the base URI of the OFREP service                                                                                   |
| WithTimeout          | Set the timeout for the http client used for communication with the OFREP service (ignored if custom client is used)    |
| WithFromEnv          | Configure the provider using environment variables                                                                      |

For example, consider below example which sets bearer token and provides a customized http client,

```go
import(
	"github.com/open-feature/go-sdk-contrib/providers/ofrep"
	of "github.com/open-feature/go-sdk/openfeature"
)

provider := ofrep.NewProvider(
    "http://localhost:8016",
    ofrep.WithBearerToken("TOKEN"),
    ofrep.WithClient(&http.Client{
        Timeout: 1 * time.Second,
    }))
```

### Environment Variable Configuration (Experimental)

You can use the `WithFromEnv()` option to configure the provider using environment variables:

```go
provider := ofrep.NewProvider(
    "http://localhost:8016",
    ofrep.WithFromEnv())
```

Supported environment variables:

| Environment Variable | Description                                                                           | Example                   |
| -------------------- | ------------------------------------------------------------------------------------- | ------------------------- |
| OFREP_ENDPOINT       | Base URI for the OFREP service (overrides the baseUri parameter)                      | `http://localhost:8016`   |
| OFREP_TIMEOUT_MS     | Timeout duration in milliseconds for HTTP requests (ignored if custom client is used) | `5000`                    |
| OFREP_HEADERS        | Comma-separated custom headers                                                        | `Key1=Value1,Key2=Value2` |
