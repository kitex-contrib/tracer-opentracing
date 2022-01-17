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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewJaegerHook(t *testing.T) {
	convey.Convey("TestNewJaegerHook", t, func() {
		convey.Convey("success", func() {
			_ = NewJaegerHook(mocktracer.New())
		})
	})
}

func Test_jaegerHook_BeforeProcess(t *testing.T) {
	convey.Convey("Test_jaegerHook_BeforeProcess", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewJaegerHook(nil)
			cmd := redis.NewStringCmd(ctx, "get")
			_, err := jh.BeforeProcess(ctx, cmd)
			assert.Equal(t, err, nil)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			_, err := jh.BeforeProcess(ctx, cmd)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 0)
		})
	})
}

func Test_jaegerHook_AfterProcess(t *testing.T) {
	convey.Convey("Test_jaegerHook_AfterProcess", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewJaegerHook(nil)
			cmd := redis.NewStringCmd(ctx, "get")
			err := jh.AfterProcess(ctx, cmd)
			assert.Equal(t, err, nil)
		})
		convey.Convey("extract span from ctx nil", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			err := jh.AfterProcess(ctx, cmd)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 0)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			ctx, err := jh.BeforeProcess(ctx, cmd)
			assert.Equal(t, err, nil)
			err = jh.AfterProcess(ctx, cmd)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
		convey.Convey("success and cmd err", func() {
			patche := gomonkey.ApplyFuncSeq(isRedisError, []gomonkey.OutputCell{
				{Values: gomonkey.Params{true}},
			})
			defer patche.Reset()

			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewBoolResult(false, redis.ErrClosed)
			ctx, err := jh.BeforeProcess(ctx, cmd)
			assert.Equal(t, err, nil)
			err = jh.AfterProcess(ctx, cmd)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
	})
}

func Test_jaegerHook_BeforeProcessPipeline(t *testing.T) {
	convey.Convey("Test_jaegerHook_BeforeProcessPipeline", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewJaegerHook(nil)
			cmd := redis.NewStringCmd(ctx, "get")
			_, err := jh.BeforeProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			_, err := jh.BeforeProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 0)
		})
	})
}

func Test_jaegerHook_AfterProcessPipeline(t *testing.T) {
	convey.Convey("Test_jaegerHook_AfterProcessPipeline", t, func() {
		convey.Convey("Tracer nil", func() {
			ctx := context.Background()
			jh := NewJaegerHook(nil)
			cmd := redis.NewStringCmd(ctx, "get")
			err := jh.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
		})
		convey.Convey("extract span from ctx nil", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			err := jh.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 0)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewStringCmd(ctx, "get")
			ctx, err := jh.BeforeProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
			err = jh.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
		convey.Convey("success and cmd err", func() {
			patche := gomonkey.ApplyFuncSeq(isRedisError, []gomonkey.OutputCell{
				{Values: gomonkey.Params{true}},
			})
			defer patche.Reset()

			ctx := context.Background()
			tracer := mocktracer.New()
			jh := NewJaegerHook(tracer)
			cmd := redis.NewBoolResult(false, redis.ErrClosed)
			ctx, err := jh.BeforeProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
			err = jh.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
	})
}

func Test_jaegerHook_getPipeLineLogKey(t *testing.T) {
	convey.Convey("Test_jaegerHook_getPipeLineLogKey", t, func() {
		convey.Convey("success", func() {
			assert.Equal(t, (&jaegerHook{}).getPipeLineLogKey("a", 1), "a-1")
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
