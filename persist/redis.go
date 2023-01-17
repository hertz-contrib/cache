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
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStore store http response in redis
type RedisStore struct {
	RedisClient *redis.Client
}

// NewRedisStore create a redis memory store with redis client
func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		RedisClient: redisClient,
	}
}

// Set put key value pair to redis, and expire after expireDuration
func (store *RedisStore) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	payload, err := Serialize(value)
	if err != nil {
		return err
	}

	return store.RedisClient.Set(ctx, key, payload, expire).Err()
}

// Delete remove key in redis, do nothing if key doesn't exist
func (store *RedisStore) Delete(ctx context.Context, key string) error {
	return store.RedisClient.Del(ctx, key).Err()
}

// Get retrieves an item from redis, if key doesn't exist, return ErrCacheMiss
func (store *RedisStore) Get(ctx context.Context, key string, value interface{}) error {
	payload, err := store.RedisClient.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}

	if err != nil {
		return err
	}
	return Deserialize(payload, value)
}
