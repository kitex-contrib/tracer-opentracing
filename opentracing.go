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
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/opentracing/opentracing-go"
)

type commonTracer struct {
	tracer            opentracing.Tracer
	formOperationName func(context.Context) string
}

func setCommonTag(span opentracing.Span, st rpcinfo.RPCStats) {
	span.SetTag("read_cost", uint64(utils.CalculateEventCost(st, stats.ReadStart, stats.ReadFinish).Microseconds()))
	span.SetTag("wait_read_cost", uint64(utils.CalculateEventCost(st, stats.WaitReadStart, stats.WaitReadFinish).Milliseconds()))
	span.SetTag("write_cost", uint64(utils.CalculateEventCost(st, stats.WriteStart, stats.WriteFinish).Milliseconds()))
	span.SetTag("send_size", st.SendSize())
	span.SetTag("recv_size", st.RecvSize())
}
