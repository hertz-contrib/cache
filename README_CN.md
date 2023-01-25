# cache (这是一个社区驱动的项目)
[English](README.md) | 中文

这是 Hertz 的一个中间件。

是一个用于缓存响应的 Hertz 中间件，支持 multi-backend。

- [memory](#memory)
- [redis](#redis)

这个仓库是从 [gin-cache](https://github.com/chenyahui/gin-cache) fork 而来的，并为 hertz 进行了适配。

## 使用方法

### 开始使用

如何下载并安装它：

```bash
go get github.com/hertz-contrib/cache
```

如何导入进你的代码：

```go
import "github.com/hertz-contrib/cache"
```

## 示例代码

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

## 许可证

本项目采用Apache许可证。参见 [LICENSE-APACHE](LICENSE-APACHE) 文件中的完整许可证文本。