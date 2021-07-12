// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opentracing

import (
	"context"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/opentracing/opentracing-go"
)

var _ stats.Tracer = &clientTracer{}

type clientTracer struct {
	commonTracer
}

func (o *clientTracer) Start(ctx context.Context) context.Context {
	var operationName string
	if o.formOperationName != nil {
		operationName = o.formOperationName(ctx)
	}
	_, ctx = opentracing.StartSpanFromContextWithTracer(ctx, o.tracer, operationName)
	return ctx
}

func (o *clientTracer) Finish(ctx context.Context) {
	span := opentracing.SpanFromContext(ctx)

	ri := rpcinfo.GetRPCInfo(ctx)
	st := ri.Stats()
	// set common rpc tag
	setCommonTag(span, st)
	// set client connection cost tag
	span.SetTag("conn_cost", uint64(utils.CalculateEventCost(st, stats.ClientConnStart, stats.ClientConnFinish).Milliseconds()))

	span.Finish()
}

// DefaultClientOption return client option with opentracing global tracer and default operation name formater.
func DefaultClientOption() client.Option {
	return ClientOption(opentracing.GlobalTracer(), func(ctx context.Context) string {
		endpoint := rpcinfo.GetRPCInfo(ctx).From()
		return endpoint.ServiceName() + "::" + endpoint.Method()
	})
}

// ClientOption return client option with specified tracer and operation name formater.
func ClientOption(tracer opentracing.Tracer, formOperationName func(c context.Context) string) client.Option {
	ct := &clientTracer{}
	ct.tracer = tracer
	ct.formOperationName = formOperationName
	return client.WithTracer(ct)
}
