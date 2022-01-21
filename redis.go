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
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracerLog "github.com/opentracing/opentracing-go/log"
)

const (
	operationRedis = "Redis-"
	logCmdName     = "command"
	logCmdArgs     = "args"
	logCmdResult   = "result"
)

type contextKey int

const (
	cmdStart contextKey = iota
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

// BeforeProcess redis before execute action do something
func (rh *redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if rh.tracer == nil {
		return ctx, nil
	}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, rh.tracer, operationRedis+cmd.Name())

	ctx = context.WithValue(ctx, cmdStart, span)
	return ctx, nil
}

// AfterProcess redis after execute action do something
func (rh *redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if rh.tracer == nil {
		return nil
	}
	span, ok := ctx.Value(cmdStart).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	span.LogFields(tracerLog.String(logCmdName, cmd.Name()))
	span.LogFields(tracerLog.Object(logCmdArgs, cmd.Args()))
	span.LogFields(tracerLog.Object(logCmdResult, cmd.String()))

	if err := cmd.Err(); isRedisError(err) {
		span.LogFields(tracerLog.Error(err))
		span.SetTag(string(ext.Error), true)
	}

	return nil
}

// BeforeProcessPipeline before command process handle
func (rh *redisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if rh.tracer == nil {
		return ctx, nil
	}

	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, rh.tracer, operationRedis+"pipeline")

	ctx = context.WithValue(ctx, cmdStart, span)

	return ctx, nil
}

// AfterProcessPipeline after command process handle
func (rh *redisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if rh.tracer == nil {
		return nil
	}

	span, ok := ctx.Value(cmdStart).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	hasErr := false
	for idx, cmd := range cmds {
		if err := cmd.Err(); isRedisError(err) {
			hasErr = true
		}
		span.LogFields(tracerLog.String(rh.getPipeLineLogKey(logCmdName, idx), cmd.Name()))
		span.LogFields(tracerLog.Object(rh.getPipeLineLogKey(logCmdArgs, idx), cmd.Args()))
		span.LogFields(tracerLog.String(rh.getPipeLineLogKey(logCmdResult, idx), cmd.String()))
	}
	if hasErr {
		span.SetTag(string(ext.Error), true)
	}
	return nil
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
