# cache (This is a community driven project)

English | [中文](README_CN.md)

This is a middleware for hertz.

Hertz middleware for cache response management with multi-backend support:

- [memory](#memory)
- [redis](#redis)

 This repo is forked from [gin-cache](https://github.com/chenyahui/gin-cache) and adapted for hertz.

## Usage

### Start using it

Download and install it:

```bash
go get github.com/hertz-contrib/cache
```

Import it in your code:

```go
import "github.com/hertz-contrib/cache"
```

## Basic Examples

### memory

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "sync/atomic"
    "time"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cache"
	"github.com/hertz-contrib/cache/persist"
)

func main() {
    h := server.New()

    memoryStore := persist.NewMemoryStore(1 * time.Minute)

    var cacheHitCount, cacheMissCount int32

    h.Use(cache.NewCacheByRequestURI(
        memoryStore,
        2*time.Second,
        cache.WithOnHitCache(func(ctx context.Context, c *app.RequestContext) {
            atomic.AddInt32(&cacheHitCount, 1)
        }),
        cache.WithOnMissCache(func(ctx context.Context, c *app.RequestContext) {
            atomic.AddInt32(&cacheMissCount, 1)
        }),
    ))
    h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
        c.String(http.StatusOK, "hello world")
    })
    h.GET("/get_hit_count", func(ctx context.Context, c *app.RequestContext) {
        c.String(200, fmt.Sprintf("total hit count: %d", cacheHitCount))
    })
    h.GET("/get_miss_count", func(ctx context.Context, c *app.RequestContext) {
        c.String(200, fmt.Sprintf("total miss count: %d", cacheMissCount))
    })

    h.Spin()
}
```

### redis

```go
package main

import (
    "context"
    "net/http"
    "time"
    
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/go-redis/redis/v8"
	"github.com/hertz-contrib/cache"
	"github.com/hertz-contrib/cache/persist"
)

func main() {
    h := server.New()

    redisStore := persist.NewRedisStore(redis.NewClient(&redis.Options{
        Network: "tcp",
        Addr:    "127.0.0.1:6379",
    }))

    h.Use(cache.NewCacheByRequestURI(redisStore, 2*time.Second))
    h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
        c.String(http.StatusOK, "hello world")
    })
    h.Spin()
}
```

## License

This project is under Apache License. See the [LICENSE-APACHE](LICENSE-APACHE) file for the full license text.