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
	"encoding/gob"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/cache/persist"
	"golang.org/x/sync/singleflight"
)

// Strategy the cache strategy
type Strategy struct {
	CacheKey string

	// CacheStore if nil, use default cache store instead
	CacheStore persist.CacheStore

	// CacheDuration
	CacheDuration time.Duration
}

// GetCacheStrategyByRequest User can use this function to design custom cache strategy by request.
// The first return value bool means whether this request should be cached.
// The second return value Strategy determine the special strategy by this request.
type GetCacheStrategyByRequest func(ctx context.Context, c *app.RequestContext) (bool, Strategy)

const (
	errMissingCacheStrategy                  = "[CACHE] cache strategy is nil"
	getCacheErrorFormat                      = "[CACHE] get cache error: %s, cache key: %s"
	setCacheKeyErrorFormat                   = "[CACHE] set cache key error: %s, cache key: %s"
	getRequestUriIgnoreQueryOrderErrorFormat = "[CACHE] getRequestUriIgnoreQueryOrder error: %s"
	writeResponseErrorFormat                 = "[CACHE] write response error: %s"
	singleFlightErrorFormat                  = "[CACHE] call the function in-flight error: %s"
	fallbackCacheKeyFormat                   = "[CACHE] Fallback to default cache key: %s"
)

// NewCache user must pass getCacheKey to describe the way to generate cache key
func NewCache(
	defaultCacheStore persist.CacheStore,
	defaultExpire time.Duration,
	opts ...Option,
) app.HandlerFunc {
	options := newOptions(opts...)
	return newCache(defaultCacheStore, defaultExpire, options)
}

func newCache(
	defaultCacheStore persist.CacheStore,
	defaultExpire time.Duration,
	options *Options,
) app.HandlerFunc {
	if options.getCacheStrategyByRequest == nil {
		panic(errMissingCacheStrategy)
	}

	sfGroup := singleflight.Group{}

	return func(ctx context.Context, c *app.RequestContext) {
		shouldCache, cacheStrategy := options.getCacheStrategyByRequest(ctx, c)
		if !shouldCache {
			c.Next(ctx)
			return
		}

		cacheKey := cacheStrategy.CacheKey

		if options.prefixKey != "" {
			cacheKey = options.prefixKey + cacheKey
		}

		// merge options
		cacheStore := defaultCacheStore
		if cacheStrategy.CacheStore != nil {
			cacheStore = cacheStrategy.CacheStore
		}

		cacheDuration := defaultExpire
		if cacheStrategy.CacheDuration > 0 {
			cacheDuration = cacheStrategy.CacheDuration
		}

		// read cache first
		{
			respCache := &ResponseCache{}
			err := cacheStore.Get(ctx, cacheKey, &respCache)
			if err == nil {
				replyWithCache(ctx, c, options, respCache)
				options.hitCacheCallback(ctx, c)
				return
			}

			if !errors.Is(err, persist.ErrCacheMiss) {
				hlog.CtxErrorf(ctx, getCacheErrorFormat, err, cacheKey)
			}
			options.missCacheCallback(ctx, c)
		}

		// cache miss, then call the backend

		// use responseCacheWriter in order to record the response
		cacheWriter := &responseCacheWriter{
			Response: &c.Response,
		}

		inFlight := false
		rawRespCache, err, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
			if options.singleFlightForgetTimeout > 0 {
				forgetTimer := time.AfterFunc(options.singleFlightForgetTimeout, func() {
					sfGroup.Forget(cacheKey)
				})
				defer forgetTimer.Stop()
			}

			c.Next(ctx)

			inFlight = true

			respCache := &ResponseCache{}
			respCache.fillWithCacheWriter(cacheWriter, options.withoutHeader)

			// only cache 2xx response
			if !c.IsAborted() && cacheWriter.StatusCode() < 300 && cacheWriter.StatusCode() >= 200 {
				if err := cacheStore.Set(ctx, cacheKey, respCache, cacheDuration); err != nil {
					hlog.CtxErrorf(ctx, setCacheKeyErrorFormat, err, cacheKey)
				}
			}

			return respCache, nil
		})

		if err != nil {
			hlog.CtxErrorf(ctx, singleFlightErrorFormat, err)
		}

		if !inFlight {
			replyWithCache(ctx, c, options, rawRespCache.(*ResponseCache))
			options.shareSingleFlightCallback(ctx, c)
		}
	}
}

// KeyStrategy defines the interface for cache key generation strategies.
type KeyStrategy interface {
	GenerateKey(c *app.RequestContext) (string, error)
}

// ByURI implements KeyStrategy using the request URI.
type ByURI struct{}

func (s *ByURI) GenerateKey(c *app.RequestContext) (string, error) {
	return string(c.Request.RequestURI()), nil
}

// ByURIWithIgnoreQueryOrder implements KeyStrategy using the request URI with ordered query parameters.
type ByURIWithIgnoreQueryOrder struct{}

func (s *ByURIWithIgnoreQueryOrder) GenerateKey(c *app.RequestContext) (string, error) {
	return getRequestUriIgnoreQueryOrder(string(c.Request.RequestURI()))
}

// ByPath implements KeyStrategy using the request path.
type ByPath struct{}

func (s *ByPath) GenerateKey(c *app.RequestContext) (string, error) {
	return b2s(c.Request.Path()), nil
}

// NewCacheByKeyStrategy is a shortcut function for caching responses based on configurable key generation strategies.
func NewCacheByKeyStrategy(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, strategy KeyStrategy, opts ...Option) app.HandlerFunc {
	cacheStrategy := func(ctx context.Context, c *app.RequestContext) (bool, Strategy) {
		cacheKey, err := strategy.GenerateKey(c)
		if err != nil {
			hlog.CtxErrorf(ctx, getRequestUriIgnoreQueryOrderErrorFormat, err)
			cacheKey = string(c.Request.RequestURI())
			hlog.CtxErrorf(ctx, fallbackCacheKeyFormat, err)
		}
		return true, Strategy{
			CacheKey: cacheKey,
		}
	}

	var options []Option
	options = append(options, WithCacheStrategyByRequest(cacheStrategy))
	options = append(options, opts...)

	return NewCache(defaultCacheStore, defaultExpire, options...)
}

// NewCacheByRequestURI a shortcut function for caching response by uri.
func NewCacheByRequestURI(store persist.CacheStore, duration time.Duration, opts ...Option) app.HandlerFunc {
	strategy := &ByURI{}
	return NewCacheByKeyStrategy(store, duration, strategy, opts...)
}

// NewCacheByRequestURIWithIgnoreQueryOrder a shortcut function for caching response by uri and ignore query param order.
func NewCacheByRequestURIWithIgnoreQueryOrder(store persist.CacheStore, duration time.Duration, opts ...Option) app.HandlerFunc {
	strategy := &ByURIWithIgnoreQueryOrder{}
	return NewCacheByKeyStrategy(store, duration, strategy, opts...)
}

// NewCacheByRequestPath a shortcut function for caching response by url path, means will discard the query params.
func NewCacheByRequestPath(store persist.CacheStore, duration time.Duration, opts ...Option) app.HandlerFunc {
	strategy := &ByPath{}
	return NewCacheByKeyStrategy(store, duration, strategy, opts...)
}

// getRequestUriIgnoreQueryOrder returns a URI with query parameters sorted alphabetically by key and value.
func getRequestUriIgnoreQueryOrder(requestURI string) (string, error) {
	parsedUrl, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return "", err
	}

	values := parsedUrl.Query()

	if len(values) == 0 {
		return requestURI, nil
	}

	queryKeys := make([]string, 0, len(values))
	for queryKey := range values {
		queryKeys = append(queryKeys, queryKey)
	}
	sort.Strings(queryKeys)

	queryVals := make([]string, 0, len(values))
	for _, queryKey := range queryKeys {
		sort.Strings(values[queryKey])
		for _, val := range values[queryKey] {
			queryVals = append(queryVals, queryKey+"="+val)
		}
	}

	return parsedUrl.Path + "?" + strings.Join(queryVals, "&"), nil
}

func init() {
	gob.Register(&ResponseCache{})
}

// ResponseCache record the http response cache
type ResponseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

func (c *ResponseCache) fillWithCacheWriter(cacheWriter *responseCacheWriter, withoutHeader bool) {
	c.Status = cacheWriter.StatusCode()
	body := cacheWriter.Body()
	buf := make([]byte, len(body))
	copy(buf, body)
	c.Data = buf
	if !withoutHeader {
		c.Header = make(map[string][]string)
		cacheWriter.Header.VisitAll(func(key, value []byte) {
			if c.Header.Get(b2s(key)) != "" {
				c.Header.Add(b2s(key), b2s(value))
			} else {
				c.Header.Set(b2s(key), b2s(value))
			}
		})
	}
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// responseCacheWriter
type responseCacheWriter struct {
	*protocol.Response
}

func replyWithCache(
	ctx context.Context,
	c *app.RequestContext,
	options *Options,
	respCache *ResponseCache,
) {
	options.beforeReplyWithCacheCallback(c, respCache)

	c.Response.SetStatusCode(respCache.Status)

	if !options.withoutHeader {
		for key, values := range respCache.Header {
			for _, val := range values {
				c.Response.Header.Set(key, val)
			}
		}
	}

	if _, err := c.Response.BodyWriter().Write(respCache.Data); err != nil {
		hlog.CtxErrorf(ctx, writeResponseErrorFormat, err)
	}

	// abort handler chain and return directly
	c.Abort()
}
