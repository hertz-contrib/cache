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
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestOptions(t *testing.T) {
	options := Options{
		getCacheStrategyByRequest: func(ctx context.Context, c *app.RequestContext) (bool, Strategy) {
			return true, Strategy{
				CacheKey: "test-key1",
			}
		},
		hitCacheCallback:             defaultHitCacheCallback,
		missCacheCallback:            defaultMissCacheCallback,
		beforeReplyWithCacheCallback: defaultBeforeReplyWithCacheCallback,
		shareSingleFlightCallback:    defaultShareSingleFlightCallback,
		singleFlightForgetTimeout:    1 * time.Second,
		ignoreQueryOrder:             false,
		prefixKey:                    "prefix1",
		withoutHeader:                false,
	}

	w, x, y, z := "", "", "", ""

	f, strategy := options.getCacheStrategyByRequest(nil, nil)
	assert.DeepEqual(t, "test-key1", strategy.CacheKey)
	assert.True(t, f)
	options.hitCacheCallback(nil, nil)
	assert.DeepEqual(t, "", w)
	options.missCacheCallback(nil, nil)
	assert.DeepEqual(t, "", x)
	options.beforeReplyWithCacheCallback(nil, nil)
	assert.DeepEqual(t, "", y)
	options.shareSingleFlightCallback(nil, nil)
	assert.DeepEqual(t, "", z)
	assert.DeepEqual(t, 1*time.Second, options.singleFlightForgetTimeout)
	assert.DeepEqual(t, "prefix1", options.prefixKey)
	assert.False(t, options.ignoreQueryOrder)
	assert.False(t, options.withoutHeader)

	opts := make([]Option, 0)
	opts = append(opts,
		WithCacheStrategyByRequest(func(ctx context.Context, c *app.RequestContext) (bool, Strategy) {
			return true, Strategy{
				CacheKey: "test-key2",
			}
		}),
		WithOnHitCache(func(c context.Context, ctx *app.RequestContext) {
			w = "W"
		}),
		WithOnMissCache(func(c context.Context, ctx *app.RequestContext) {
			x = "X"
		}),
		WithBeforeReplyWithCache(func(c *app.RequestContext, cache *ResponseCache) {
			y = "Y"
		}),
		WithSingleFlightForgetTimeout(2*time.Second),
		WithOnShareSingleFlight(func(ctx context.Context, c *app.RequestContext) {
			z = "Z"
		}),
		WithIgnoreQueryOrder(true),
		WithoutHeader(true),
		WithPrefixKey("prefix2"),
	)

	options.Apply(opts)

	f, strategy = options.getCacheStrategyByRequest(nil, nil)
	assert.DeepEqual(t, "test-key2", strategy.CacheKey)
	assert.True(t, f)
	options.hitCacheCallback(nil, nil)
	assert.DeepEqual(t, "W", w)
	options.missCacheCallback(nil, nil)
	assert.DeepEqual(t, "X", x)
	options.beforeReplyWithCacheCallback(nil, nil)
	assert.DeepEqual(t, "Y", y)
	options.shareSingleFlightCallback(nil, nil)
	assert.DeepEqual(t, "Z", z)
	assert.DeepEqual(t, 2*time.Second, options.singleFlightForgetTimeout)
	assert.DeepEqual(t, "prefix2", options.prefixKey)
	assert.True(t, options.ignoreQueryOrder)
	assert.True(t, options.withoutHeader)
}
