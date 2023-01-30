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

package persist

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/jellydator/ttlcache/v2"
)

const (
	setTTLErrorFormat = "[CACHE] set ttl for memory store error: %s"
)

// MemoryStore local memory cache store
type MemoryStore struct {
	Cache *ttlcache.Cache
}

// NewMemoryStore allocate a local memory store with default expiration
func NewMemoryStore(defaultExpiration time.Duration) *MemoryStore {
	cacheStore := ttlcache.NewCache()
	if err := cacheStore.SetTTL(defaultExpiration); err != nil {
		hlog.Errorf(setTTLErrorFormat, err)
	}

	// disable SkipTTLExtensionOnHit default
	cacheStore.SkipTTLExtensionOnHit(true)

	return &MemoryStore{
		Cache: cacheStore,
	}
}

// Set put key value pair to memory store, and expire after expireDuration
func (c *MemoryStore) Set(ctx context.Context, key string, value interface{}, expireDuration time.Duration) error {
	return c.Cache.SetWithTTL(key, value, expireDuration)
}

// Delete remove key in memory store, do nothing if key doesn't exist
func (c *MemoryStore) Delete(ctx context.Context, key string) error {
	return c.Cache.Remove(key)
}

// Get get key in memory store, if key doesn't exist, return ErrCacheMiss
func (c *MemoryStore) Get(ctx context.Context, key string, value interface{}) error {
	val, err := c.Cache.Get(key)
	if errors.Is(err, ttlcache.ErrNotFound) {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	v.Elem().Set(reflect.ValueOf(val))
	return nil
}
