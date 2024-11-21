# OpenFeature Remote Evaluation Protocol Provider

This is a fork of `open-feature/go-sdk-contrib/pkg/providers/ofrep@0.1.5` and the **opinionated** Go implementation of the OFREP provider.

The package contains two providers, which can be used to interact with the OFREP single and bulk flag evaluation endpoint. The goal is to support the drop-in replacement of the existing OFREP GO provider why waiting for official implementation.

## Usage
Add a line to your `go.mod` file

```
replace github.com/open-feature/go-sdk-contrib/providers/ofrep => github.com/erka/openfeature-go-ofrep-provider v0.0.1
```

For OFREP Bulk provider example

```go
import(
	"github.com/open-feature/go-sdk-contrib/providers/ofrep"
	of "github.com/open-feature/go-sdk/openfeature"
)
///.....
of.SetEvaluationContext(of.NewEvaluationContext(...))
p := ofrep.NewBulkProvider(baseUrl, ....)
err := of.SetProviderAndWait(p)
if err != nil {
  // handle the error
}
client := of.NewClient("app")
```
