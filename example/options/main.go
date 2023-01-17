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

package main

import (
	"cache"
	"cache/persist"
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
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
