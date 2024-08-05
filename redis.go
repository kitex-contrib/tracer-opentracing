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
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracerLog "github.com/opentracing/opentracing-go/log"
	"github.com/redis/go-redis/v9"
)

const (
	operationRedis = "Redis-"
	logCmdName     = "command"
	logCmdArgs     = "args"
	logCmdResult   = "result"
	redisPipeline  = "pipeline"
	redisDial      = "dial"
)

// redisHook implements go-redis hook
type redisHook struct {
	tracer opentracing.Tracer
}

// NewRedisHook return redis.Hook
func NewRedisHook(tracer opentracing.Tracer) redis.Hook {
	return &redisHook{
		tracer: tracer,
	}
}

func (rh *redisHook) DialHook(hook redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if rh.tracer == nil {
			return hook(ctx, network, addr)
		}
		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, rh.tracer, operationRedis+redisDial)
		conn, err := hook(ctx, network, addr)
		defer span.Finish()

		if isRedisError(err) {
			span.LogFields(tracerLog.Error(err))
			span.SetTag(string(ext.Error), true)
		}

		return conn, err
	}
}

func (rh *redisHook) ProcessHook(hook redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if rh.tracer == nil {
			return hook(ctx, cmd)
		}
		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, rh.tracer, operationRedis+cmd.Name())
		err := hook(ctx, cmd)
		defer span.Finish()

		span.LogFields(tracerLog.String(logCmdName, cmd.Name()))
		span.LogFields(tracerLog.Object(logCmdArgs, cmd.Args()))
		span.LogFields(tracerLog.Object(logCmdResult, cmd.String()))
		if isRedisError(err) {
			span.LogFields(tracerLog.Error(err))
			span.SetTag(string(ext.Error), true)
		}

		return err
	}
}

func (rh *redisHook) ProcessPipelineHook(hook redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if rh.tracer == nil {
			return hook(ctx, cmds)
		}
		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, rh.tracer, operationRedis+redisPipeline)
		err := hook(ctx, cmds)
		defer span.Finish()
		for idx, cmd := range cmds {
			span.LogFields(tracerLog.String(rh.getPipeLineLogKey(logCmdName, idx), cmd.Name()))
			span.LogFields(tracerLog.Object(rh.getPipeLineLogKey(logCmdArgs, idx), cmd.Args()))
			span.LogFields(tracerLog.String(rh.getPipeLineLogKey(logCmdResult, idx), cmd.String()))
		}
		if isRedisError(err) {
			span.LogFields(tracerLog.Error(err))
			span.SetTag(string(ext.Error), true)
		}
		return err
	}
}

func (rh *redisHook) getPipeLineLogKey(logField string, idx int) string {
	return logField + "-" + strconv.Itoa(idx)
}

func isRedisError(err error) bool {
	if err == redis.Nil {
		return false
	}
	_, ok := err.(redis.Error)
	return ok
}
