# opentracing
[Opentracing](https://github.com/opentracing/opentracing-go) tracer for [Kitex](https://github.com/cloudwego/kitex)

## Server usage
```go
import (
    ...
    internal_opentracing "github.com/kitex-contrib/tracer-opentracing"
    ...
)

func main() {
    ...
    svr := echo.NewServer(new(EchoImpl), internal_opentracing.DefaultServerOption())
    ...
}
```
`DefaultServerOption` will use opentracing global tracer as tracer, and `{Service Name}::{Method Name}` as operation name. You can customize both by `ServerOption`.

## Client usage
```go
import (
    ...
    internal_opentracing "github.com/kitex-contrib/tracer-opentracing"
    ...
)

func main() {
    ...
    client, err := echo.NewClient("echo", internal_opentracing.DefaultClientOption())
	if err != nil {
		log.Fatal(err)
	}
    ...
}
```
Just like server, `DefaultClientOption` will use opentracing global tracer as tracer, and `{Service Name}::{Method Name}` as operation name. You can customize both by `ClientOption`.
## Example
[Executable Example](https://github.com/cloudwego/kitex-examples/tree/main/tracer)