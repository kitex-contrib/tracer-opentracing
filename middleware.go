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
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/opentracing/opentracing-go"
)

const (
	// SpanContextKey transit key, transited by HTTP2 will convert case, so use uppercase
	SpanContextKey = "JAEGERSPANCONTEXT"
)

type opentracingCtx int

const (
	traceContainerKey opentracingCtx = iota
)

func SpanContextInjectMW(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		span := opentracing.SpanFromContext(ctx)
		var b bytes.Buffer
		span.Tracer().Inject(span.Context(), opentracing.Binary, &b)
		ctx = metainfo.WithValue(ctx, SpanContextKey, base64.StdEncoding.EncodeToString(b.Bytes()))
		return next(ctx, req, resp)
	}
}

func SpanContextExtractMW(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) error {
		tc, ok := ctx.Value(traceContainerKey).(*traceContainer)
		if !ok {
			return errors.New("no opentracing tracer found in context")
		}
		svrTracer := tc.serverTracer
		var operationName string
		if svrTracer.formOperationName != nil {
			operationName = svrTracer.formOperationName(ctx)
		}

		var opts []opentracing.StartSpanOption

		ri := rpcinfo.GetRPCInfo(ctx)
		startTime := ri.Stats().GetEvent(stats.RPCStart).Time()
		opts = append(opts, opentracing.StartTime(startTime))

		if sck, ok := metainfo.GetValue(ctx, SpanContextKey); ok {
			binaryBytes, err := base64.StdEncoding.DecodeString(sck)
			if err != nil {
				return fmt.Errorf("decode opentracing binary failed, %w", err)
			}
			parentContext, err := svrTracer.tracer.Extract(opentracing.Binary, bytes.NewBuffer(binaryBytes))
			if err != nil {
				return fmt.Errorf("extract SpanContext failed, %w", err)
			}
			opts = append(opts, opentracing.ChildOf(parentContext))
		}

		rpcSpan, ctx := opentracing.StartSpanFromContextWithTracer(ctx, svrTracer.tracer, operationName, opts...)
		tc.span = rpcSpan
		return next(ctx, req, resp)
	}
}
