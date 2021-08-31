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
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
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
	ri := rpcinfo.GetRPCInfo(ctx)
	startTime := ri.Stats().GetEvent(stats.RPCStart).Time()
	_, ctx = opentracing.StartSpanFromContextWithTracer(ctx, o.tracer, operationName, opentracing.StartTime(startTime))
	return ctx
}

func (o *clientTracer) Finish(ctx context.Context) {
	rpcSpan := opentracing.SpanFromContext(ctx)

	ri := rpcinfo.GetRPCInfo(ctx)
	st := ri.Stats()

	// new common rpc span
	o.newCommonSpan(rpcSpan, st)
	// new establish connection span
	o.newEventSpan("establish connection", st, stats.ClientConnStart, stats.ClientConnFinish, rpcSpan.Context())

	rpcSpan.FinishWithOptions(opentracing.FinishOptions{FinishTime: st.GetEvent(stats.RPCFinish).Time()})
}

// clientOption return client option with specified tracer and operation name formater.
func clientOption(tracer opentracing.Tracer, formOperationName func(c context.Context) string) client.Option {
	ct := &clientTracer{}
	ct.tracer = tracer
	ct.formOperationName = formOperationName
	return client.WithTracer(ct)
}

func NewDefaultClientSuite() client.Suite {
	return &clientSuite{opentracing.GlobalTracer(), func(ctx context.Context) string {
		endpoint := rpcinfo.GetRPCInfo(ctx).From()
		return endpoint.ServiceName() + "::" + endpoint.Method()
	}}
}

func NewClientSuite(tracer opentracing.Tracer, formOperationName func(c context.Context) string) client.Suite {
	return &clientSuite{tracer, formOperationName}
}

type clientSuite struct {
	tracer            opentracing.Tracer
	formOperationName func(c context.Context) string
}

func (c *clientSuite) Options() []client.Option {
	var options []client.Option
	options = append(options, clientOption(c.tracer, c.formOperationName))
	options = append(options, client.WithMiddleware(SpanContextInjectMW))
	options = append(options, client.WithTransportProtocol(transport.TTHeader))
	options = append(options, client.WithMetaHandler(transmeta.ClientTTHeaderHandler))
	return options
}
