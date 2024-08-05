// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opentracing

import (
	"context"
	"net"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/redis/go-redis/v9"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisHook(t *testing.T) {
	convey.Convey("TestNewRedisHook", t, func() {
		convey.Convey("success", func() {
			_ = NewRedisHook(mocktracer.New())
		})
	})
}

func Test_redisHook_ProcessHook(t *testing.T) {
	convey.Convey("Test_redisHook_ProcessHook", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewRedisHook(nil)
			cmd := redis.NewStringCmd(ctx, "get")
			hook := jh.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
				return nil
			})
			assert.Equal(t, hook(ctx, cmd), nil)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewRedisHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			hook := jh.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
				return nil
			})
			assert.Equal(t, hook(ctx, cmd), nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
		convey.Convey("success and cmd err", func() {
			patche := gomonkey.ApplyFuncSeq(isRedisError, []gomonkey.OutputCell{
				{Values: gomonkey.Params{true}},
			})
			defer patche.Reset()

			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewRedisHook(tracer)
			cmd := redis.NewBoolResult(false, redis.ErrClosed)
			hook := jh.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
				return nil
			})
			assert.Equal(t, hook(ctx, cmd), nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
	})
}

func Test_redisHook_ProcessPipelineHook(t *testing.T) {
	convey.Convey("Test_redisHook_ProcessPipelineHook", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewRedisHook(nil)
			cmd := redis.NewStringCmd(ctx, "get")
			hook := jh.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
				return nil
			})
			assert.Equal(t, hook(ctx, []redis.Cmder{cmd}), nil)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewRedisHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			hook := jh.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
				return nil
			})
			assert.Equal(t, hook(ctx, []redis.Cmder{cmd}), nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
		convey.Convey("success and cmd err", func() {
			patche := gomonkey.ApplyFuncSeq(isRedisError, []gomonkey.OutputCell{
				{Values: gomonkey.Params{true}},
			})
			defer patche.Reset()

			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewRedisHook(tracer)
			cmd := redis.NewBoolResult(false, redis.ErrClosed)
			hook := jh.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
				return nil
			})
			assert.Equal(t, hook(ctx, []redis.Cmder{cmd}), nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
	})
}

func Test_redisHook_DialHook(t *testing.T) {
	convey.Convey("Test_redisHook_DialHook", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewRedisHook(nil)
			hook := jh.DialHook(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return nil, nil
			})
			conn, err := hook(ctx, "", "")
			assert.Equal(t, conn, nil)
			assert.Equal(t, err, nil)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewRedisHook(tracer)
			hook := jh.DialHook(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return nil, nil
			})
			conn, err := hook(ctx, "", "")
			assert.Equal(t, conn, nil)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
		convey.Convey("success and dial err", func() {
			patche := gomonkey.ApplyFuncSeq(isRedisError, []gomonkey.OutputCell{
				{Values: gomonkey.Params{true}},
			})
			defer patche.Reset()

			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewRedisHook(tracer)
			hook := jh.DialHook(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return nil, nil
			})
			conn, err := hook(ctx, "", "")
			assert.Equal(t, conn, nil)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
	})
}

func Test_redisHook_getPipeLineLogKey(t *testing.T) {
	convey.Convey("Test_redisHook_getPipeLineLogKey", t, func() {
		convey.Convey("success", func() {
			assert.Equal(t, (&redisHook{}).getPipeLineLogKey("a", 1), "a-1")
		})
	})
}

func Test_isRedisError(t *testing.T) {
	convey.Convey("Test_isRedisError", t, func() {
		convey.Convey("redis.Nil", func() {
			assert.Equal(t, isRedisError(redis.Nil), false)
		})
		convey.Convey("not redis.Nil", func() {
			assert.Equal(t, isRedisError(nil), false)
		})
	})
}
