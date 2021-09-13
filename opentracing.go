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

// Package opentracing implements KiteX tracer with opentracing.
package opentracing

import (
	"context"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/opentracing/opentracing-go"
)

type commonTracer struct {
	tracer            opentracing.Tracer
	formOperationName func(context.Context) string
}

func (c *commonTracer) newCommonSpan(span opentracing.Span, st rpcinfo.RPCStats) {
	readSpan := c.newEventSpan("read", st, stats.ReadStart, stats.ReadFinish, span.Context())
	readSpan.SetTag("recv_size", st.RecvSize())

	c.newEventSpan("wait_read", st, stats.WaitReadStart, stats.WaitReadFinish, span.Context())
	writeSpan := c.newEventSpan("write", st, stats.WriteStart, stats.WriteFinish, span.Context())
	writeSpan.SetTag("send_size", st.SendSize())
}

func (c *commonTracer) newEventSpan(operationName string, st rpcinfo.RPCStats, start, end stats.Event, parentContext opentracing.SpanContext) opentracing.Span {
	var opts []opentracing.StartSpanOption
	event := st.GetEvent(start)
	if event == nil {
		return nil
	}
	startTime := opentracing.StartTime(event.Time())
	opts = append(opts, opentracing.StartTime(startTime))
	if parentContext != nil {
		opts = append(opts, opentracing.ChildOf(parentContext))
	}
	span := c.tracer.StartSpan(operationName, opts...)
	span.FinishWithOptions(opentracing.FinishOptions{FinishTime: st.GetEvent(end).Time()})
	return span
}
