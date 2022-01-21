# opentracing
[Opentracing](https://github.com/opentracing/opentracing-go) tracer for [Kitex](https://github.com/cloudwego/kitex)

## Server usage
```go
import (
    ...
    "github.com/cloudwego/kitex/server"
    internal_opentracing "github.com/kitex-contrib/tracer-opentracing"
    ...
)

func main() {
    ...
    tracer := internal_opentracing.NewDefaultServerSuite()
    svr := echo.NewServer(new(EchoImpl), server.WithSuite(tracer))
    ...
}
```
`DefaultServerOption` will use opentracing global tracer as tracer, and `{Service Name}::{Method Name}` as operation name. You can customize both by `ServerOption`.

## Client usage
```go
import (
    ...
    "github.com/cloudwego/kitex/client"
    internal_opentracing "github.com/kitex-contrib/tracer-opentracing"
    ...
)

func main() {
    ...
    tracer := internal_opentracing.NewDefaultClientSuite()
    client, err := echo.NewClient("echo", client.WithSuite(tracer))
	if err != nil {
		log.Fatal(err)
	}
    ...
}
```
Just like server, `DefaultClientOption` will use opentracing global tracer as tracer, and `{Service Name}::{Method Name}` as operation name. You can customize both by `ClientOption`.
## Example
[Executable Example](https://github.com/cloudwego/kitex-examples/tree/main/tracer)

## Supported Components
## Redis
You need to add the hook provided by tracer-opentracing to instrument Redis client:
```go
import (
    ...
    "github.com/go-redis/redis/v8"
    internal_opentracing "github.com/kitex-contrib/tracer-opentracing"
    ...
)

func main() {
    ...
    rdb := redis.NewClient(&redis.Options{...})
    rdb.AddHook(internal_opentracing.NewTracingHook())
    ...
}
```

