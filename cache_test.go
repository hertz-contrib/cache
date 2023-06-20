/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * The MIT License (MIT)
 *
 * Copyright (c) 2021 cyhone
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
* This file may have been modified by CloudWeGo authors. All CloudWeGo
* Modifications are Copyright 2022 CloudWeGo Authors.
*/

package cache

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/cache/persist"
)

func hertzHandler(middleware app.HandlerFunc, withRand bool) *route.Engine {
	r := route.NewEngine(config.NewOptions([]config.Option{}))

	r.Use(middleware)
	r.GET("/cache", func(ctx context.Context, c *app.RequestContext) {
		body := "uid:" + c.Query("uid")
		if withRand {
			body += fmt.Sprintf(",rand:%d", rand.Int())
		}
		c.String(http.StatusOK, body)
	})

	return r
}

func TestCacheByRequestPath(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cachePathMiddleware := NewCacheByRequestPath(memoryStore, 3*time.Second)

	handler := hertzHandler(cachePathMiddleware, true)

	w1 := ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	w2 := ut.PerformRequest(handler, "GET", "/cache?uid=u2", nil)
	w3 := ut.PerformRequest(handler, "GET", "/cache?uid=u3", nil)

	assert.NotEqual(t, "", w1.Body)
	assert.DeepEqual(t, w1.Body, w2.Body)
	assert.DeepEqual(t, w2.Body, w3.Body)
	assert.DeepEqual(t, w1.Code, w2.Code)
}

func TestCacheHitMissCallback(t *testing.T) {
	var cacheHitCount, cacheMissCount int32
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cachePathMiddleware := NewCacheByRequestPath(memoryStore, 3*time.Second,
		WithOnHitCache(func(ctx context.Context, c *app.RequestContext) {
			atomic.AddInt32(&cacheHitCount, 1)
		}),
		WithOnMissCache(func(ctx context.Context, c *app.RequestContext) {
			atomic.AddInt32(&cacheMissCount, 1)
		}),
	)
	handler := hertzHandler(cachePathMiddleware, true)

	ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	ut.PerformRequest(handler, "GET", "/cache?uid=u2", nil)
	ut.PerformRequest(handler, "GET", "/cache?uid=u3", nil)

	assert.DeepEqual(t, int32(2), cacheHitCount)
	assert.DeepEqual(t, int32(1), cacheMissCount)
}

func TestCacheDuration(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := NewCacheByRequestURI(memoryStore, 3*time.Second)
	handler := hertzHandler(cacheURIMiddleware, true)

	w1 := ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	time.Sleep(1 * time.Second)

	w2 := ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	assert.DeepEqual(t, w1.Body, w2.Body)
	assert.DeepEqual(t, w1.Code, w2.Code)
	time.Sleep(2 * time.Second)

	w3 := ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	assert.NotEqual(t, w1.Body, w3.Body)
}

func TestCacheByRequestURI(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := NewCacheByRequestURI(memoryStore, 3*time.Second)
	handler := hertzHandler(cacheURIMiddleware, true)

	w1 := ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	w2 := ut.PerformRequest(handler, "GET", "/cache?uid=u1", nil)
	w3 := ut.PerformRequest(handler, "GET", "/cache?uid=u2", nil)

	assert.DeepEqual(t, w1.Body, w2.Body)
	assert.DeepEqual(t, w1.Code, w2.Code)

	assert.NotEqual(t, w2.Body, w3.Body)
}

func hertzHeaderHandler(middleware app.HandlerFunc) *route.Engine {
	r := route.NewEngine(config.NewOptions([]config.Option{}))

	r.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Header("test_header_key", "test_header_value")
	})
	r.Use(middleware)
	r.GET("/cache", func(ctx context.Context, c *app.RequestContext) {
		c.Header("test_header_key", "test_header_value2")
		c.String(http.StatusOK, "value")
	})

	return r
}

func TestHeader(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := NewCacheByRequestURI(memoryStore, 3*time.Second)
	headerHandler := hertzHeaderHandler(cacheURIMiddleware)

	w1 := ut.PerformRequest(headerHandler, "GET", "/cache", nil)
	assert.DeepEqual(t, "test_header_value2", w1.Header().Get("test_header_key"))
	w2 := ut.PerformRequest(headerHandler, "GET", "/cache", nil)
	assert.DeepEqual(t, "test_header_value2", w2.Header().Get("test_header_key"))
}

func TestConcurrentRequest(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := NewCacheByRequestURI(memoryStore, 1*time.Second)
	handler := hertzHandler(cacheURIMiddleware, false)

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			uid := rand.Intn(5)
			url := fmt.Sprintf("/cache?uid=%d", uid)
			expect := fmt.Sprintf("uid:%d", uid)

			w := ut.PerformRequest(handler, "GET", url, nil)
			assert.DeepEqual(t, expect, w.Body.String())
		}()
	}

	wg.Wait()
}

func writeHeaderHandler(middleware app.HandlerFunc) *route.Engine {
	r := route.NewEngine(config.NewOptions([]config.Option{}))

	r.Use(middleware)
	r.GET("/cache", func(ctx context.Context, c *app.RequestContext) {
		c.Status(http.StatusOK)
		c.Header("hello", "world")
	})

	return r
}

func TestWriteHeader(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := NewCacheByRequestURI(memoryStore, 3*time.Second)

	handler := writeHeaderHandler(cacheURIMiddleware)
	w1 := ut.PerformRequest(handler, "GET", "/cache", nil)
	assert.DeepEqual(t, "world", w1.Header().Get("hello"))
	w2 := ut.PerformRequest(handler, "GET", "/cache", nil)
	assert.DeepEqual(t, "world", w2.Header().Get("hello"))
}

func TestGetRequestUriIgnoreQueryOrder(t *testing.T) {
	val, err := getRequestUriIgnoreQueryOrder("/test?c=3&b=2&a=1")
	assert.Nil(t, err)
	assert.DeepEqual(t, "/test?a=1&b=2&c=3", val)

	val, err = getRequestUriIgnoreQueryOrder("/test?d=4&e=5")
	assert.Nil(t, err)
	assert.DeepEqual(t, "/test?d=4&e=5", val)
}

const prefixKey = "#prefix#"

func TestPrefixKey(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := NewCacheByRequestPath(
		memoryStore,
		3*time.Second,
		WithPrefixKey(prefixKey),
	)

	requestPath := "/cache"

	handler := hertzHandler(cacheURIMiddleware, true)
	w1 := ut.PerformRequest(handler, "GET", requestPath, nil)

	err := memoryStore.Delete(context.Background(), prefixKey+requestPath)
	assert.Nil(t, err)

	w2 := ut.PerformRequest(handler, "GET", requestPath, nil)
	assert.NotEqual(t, w1.Body, w2.Body)
}

func TestNewCache_Memory(t *testing.T) {
	h := server.New(
		server.WithHostPorts("127.0.0.1:9233"))
	original := map[string][]byte{
		"/tmp-cache/ping1": []byte("{\"data\":{\"num\":1111111111}}"),
		"/tmp-cache/ping2": []byte("{\"data\":{\"num\":2222222222222222222}}"),
		"/tmp-cache/ping3": []byte("{\"data\":{\"num\":3333333333333333333333333333}}"),
	}
	h.Use(NewCache(persist.NewMemoryStore(time.Second), 3*time.Second,
		WithCacheStrategyByRequest(func(ctx context.Context, c *app.RequestContext) (bool, Strategy) {
			return true, Strategy{
				CacheKey:      c.Request.URI().String(),
				CacheDuration: 5 * time.Second,
			}
		})))
	h.GET("/tmp-cache/*path", func(ctx context.Context, c *app.RequestContext) {
		if data, ok := original[string(c.Request.Path())]; ok {
			_, _ = c.Response.BodyWriter().Write(data)
			return
		}
	})
	go h.Spin()

	tests := []struct {
		want []byte
		url  string
	}{
		{
			want: original["/tmp-cache/ping1"],
			url:  "http://127.0.0.1:9233/tmp-cache/ping1",
		},
		{
			want: original["/tmp-cache/ping2"],
			url:  "http://127.0.0.1:9233/tmp-cache/ping2",
		},
		{
			want: original["/tmp-cache/ping3"],
			url:  "http://127.0.0.1:9233/tmp-cache/ping3",
		},
	}

	for i := 0; i < 10; i++ {
		for _, tt := range tests {
			t.Run("cache data", func(t *testing.T) {
				resp, err := http.Get(tt.url)
				assert.Nil(t, err)
				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)
				got := body
				assert.DeepEqual(t, string(tt.want), string(got))
			})
		}
	}
}
