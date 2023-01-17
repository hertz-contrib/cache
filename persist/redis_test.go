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
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/go-redis/redis/v8"
)

func TestRedisStore(t *testing.T) {
	redisStore := NewRedisStore(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	}))

	ctx := context.Background()

	expectVal := "123"
	assert.Nil(t, redisStore.Set(ctx, "test", expectVal, 1*time.Second))

	value := ""
	assert.Nil(t, redisStore.Get(ctx, "test", &value))
	assert.DeepEqual(t, expectVal, value)

	time.Sleep(1 * time.Second)
	assert.DeepEqual(t, ErrCacheMiss, redisStore.Get(ctx, "test", &value))

	redisStore.Set(ctx, "test", "value", 1*time.Second)
	assert.Nil(t, redisStore.Get(ctx, "test", &value))
	assert.DeepEqual(t, "value", value)
	redisStore.Delete(ctx, "test")
	assert.DeepEqual(t, ErrCacheMiss, redisStore.Get(ctx, "test", &value))
}
