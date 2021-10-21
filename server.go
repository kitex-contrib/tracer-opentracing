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

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/opentracing/opentracing-go"
)

var _ stats.Tracer = &serverTracer{}

type serverTracer struct {
	commonTracer
}

type traceContainer struct {
	serverTracer *serverTracer
	span         opentracing.Span
}

func (o *serverTracer) Start(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, traceContainerKey, &traceContainer{serverTracer: o})
	return ctx
}

func (o *serverTracer) Finish(ctx context.Context) {
	tc, ok := ctx.Value(traceContainerKey).(*traceContainer)
	if !ok || tc.span == nil {
		panic("get tracer container failed")
	}
	rpcSpan := tc.span

	ri := rpcinfo.GetRPCInfo(ctx)
	st := ri.Stats()

	// new common rpc span
	o.newCommonSpan(rpcSpan, st)
	// new handler span
	o.newEventSpan("handler", st, stats.ServerHandleStart, stats.ServerHandleFinish, rpcSpan.Context())

	rpcSpan.FinishWithOptions(opentracing.FinishOptions{FinishTime: st.GetEvent(stats.RPCFinish).Time()})
}

// serverOption return server option with specified tracer and operation name formater.
func serverOption(tracer opentracing.Tracer, formOperationName func(c context.Context) string) server.Option {
	st := &serverTracer{}
	st.tracer = tracer
	st.formOperationName = formOperationName
	return server.WithTracer(st)
}

func NewDefaultServerSuite() server.Suite {
	return &serverSuite{opentracing.GlobalTracer(), func(ctx context.Context) string {
		endpoint := rpcinfo.GetRPCInfo(ctx).From()
		return endpoint.ServiceName() + "::" + endpoint.Method()
	}}
}

func NewServerSuite(tracer opentracing.Tracer, formOperationName func(c context.Context) string) server.Suite {
	return &serverSuite{tracer, formOperationName}
}

type serverSuite struct {
	tracer            opentracing.Tracer
	formOperationName func(c context.Context) string
}

func (c *serverSuite) Options() []server.Option {
	var options []server.Option
	options = append(options, serverOption(c.tracer, c.formOperationName))
	options = append(options, server.WithMiddleware(SpanContextExtractMW))
	options = append(options, server.WithMetaHandler(transmeta.ServerTTHeaderHandler))
	return options
}
