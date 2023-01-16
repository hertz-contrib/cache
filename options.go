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
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// Options contains all options
type Options struct {
	getCacheStrategyByRequest GetCacheStrategyByRequest

	hitCacheCallback  OnHitCacheCallback
	missCacheCallback OnMissCacheCallback

	beforeReplyWithCacheCallback BeforeReplyWithCacheCallback

	singleFlightForgetTimeout time.Duration
	shareSingleFlightCallback OnShareSingleFlightCallback

	ignoreQueryOrder bool
	prefixKey        string
	withoutHeader    bool
}

// OnHitCacheCallback define the callback when use cache
type OnHitCacheCallback app.HandlerFunc

var defaultHitCacheCallback = func(ctx context.Context, c *app.RequestContext) {}

// OnMissCacheCallback define the callback when use cache
type OnMissCacheCallback app.HandlerFunc

var defaultMissCacheCallback = func(ctx context.Context, c *app.RequestContext) {}

// OnShareSingleFlightCallback define the callback when share the singleFlight result
type OnShareSingleFlightCallback func(ctx context.Context, c *app.RequestContext)

var defaultShareSingleFlightCallback = func(ctx context.Context, c *app.RequestContext) {}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func newOptions(opts ...Option) *Options {
	options := &Options{
		hitCacheCallback:             defaultHitCacheCallback,
		missCacheCallback:            defaultMissCacheCallback,
		beforeReplyWithCacheCallback: defaultBeforeReplyWithCacheCallback,
		shareSingleFlightCallback:    defaultShareSingleFlightCallback,
	}

	options.Apply(opts)

	return options
}

// Option represents the optional function.
type Option struct {
	F func(o *Options)
}

// WithCacheStrategyByRequest set up the custom strategy by per request
func WithCacheStrategyByRequest(getGetCacheStrategyByRequest GetCacheStrategyByRequest) Option {
	return Option{
		F: func(o *Options) {
			o.getCacheStrategyByRequest = getGetCacheStrategyByRequest
		},
	}
}

// WithOnHitCache will be called when cache hit.
func WithOnHitCache(cb OnHitCacheCallback) Option {
	return Option{
		F: func(o *Options) {
			o.hitCacheCallback = cb
		},
	}
}

// WithOnMissCache will be called when cache miss.
func WithOnMissCache(cb OnMissCacheCallback) Option {
	return Option{
		F: func(o *Options) {
			o.missCacheCallback = cb
		},
	}
}

type BeforeReplyWithCacheCallback func(c *app.RequestContext, cache *ResponseCache)

var defaultBeforeReplyWithCacheCallback = func(c *app.RequestContext, cache *ResponseCache) {}

// WithBeforeReplyWithCache will be called before replying with cache.
func WithBeforeReplyWithCache(cb BeforeReplyWithCacheCallback) Option {
	return Option{
		F: func(o *Options) {
			o.beforeReplyWithCacheCallback = cb
		},
	}
}

// WithOnShareSingleFlight will be called when share the singleflight result
func WithOnShareSingleFlight(cb OnShareSingleFlightCallback) Option {
	return Option{
		F: func(o *Options) {
			o.shareSingleFlightCallback = cb
		},
	}
}

// WithSingleFlightForgetTimeout to reduce the impact of long tail requests.
// singleFlight.Forget will be called after the timeout has reached for each backend request when timeout is greater than zero.
func WithSingleFlightForgetTimeout(forgetTimeout time.Duration) Option {
	return Option{
		F: func(o *Options) {
			o.singleFlightForgetTimeout = forgetTimeout
		},
	}
}

// IgnoreQueryOrder will ignore the queries order in url when generate cache key . This option only takes effect in CacheByRequestURI function
func IgnoreQueryOrder(b bool) Option {
	return Option{
		F: func(o *Options) {
			o.ignoreQueryOrder = b
		},
	}
}

// WithPrefixKey will prefix the key
func WithPrefixKey(prefix string) Option {
	return Option{
		F: func(o *Options) {
			o.prefixKey = prefix
		},
	}
}

func WithoutHeader(b bool) Option {
	return Option{
		F: func(o *Options) {
			o.withoutHeader = b
		},
	}
}
