/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package remote

import (
	"context"
	"runtime/pprof"

	"github.com/cloudwego/kitex/pkg/profiler"
	"github.com/cloudwego/kitex/pkg/remote/transmeta"
)

const (
	keyType      = "type"
	typeTrace    = "trace"
	typeProfiler = "profiler"

	keyStage     = "stage"
	stageMeta    = "meta"
	stagePayload = "payload"

	keyFrom = "from"
)

type ProfilerController interface {
	Run(ctx context.Context) (err error)
	Stop()
	TagMeta(ctx context.Context, msg Message)
	TagPayload(ctx context.Context, msg Message)
	Untag(ctx context.Context)
}

var _ ProfilerController = (*profilerController)(nil)

func NewProfilerController(p profiler.Profiler) ProfilerController {
	return &profilerController{p: p}
}

type profilerController struct {
	p profiler.Profiler
}

func (c *profilerController) Run(ctx context.Context) (err error) {
	// tagging current goroutine
	// we could filter tag from pprof data to figure out that how much cost our profiler cause
	ctx = pprof.WithLabels(ctx, pprof.Labels(keyType, typeProfiler))
	return c.p.Run(ctx)
}

func (c *profilerController) Stop() {
	c.p.Stop()
}

func (c *profilerController) TagMeta(ctx context.Context, msg Message) {
	ti := msg.TransInfo()
	if ti == nil || ti.TransIntInfo()[transmeta.FromService] == "" {
		return
	}
	c.p.Tag(ctx,
		keyType, typeTrace,
		keyStage, stageMeta,
		keyFrom, ti.TransIntInfo()[transmeta.FromService],
	)
}

func (c *profilerController) TagPayload(ctx context.Context, msg Message) {
	ti := msg.TransInfo()
	// that's means TagMeta already set the tags
	if ti.TransIntInfo()[transmeta.FromService] != "" {
		return
	}

	ri := msg.RPCInfo()
	if ri == nil || ri.From() == nil {
		return
	}
	c.p.Tag(ctx,
		keyType, typeTrace,
		keyStage, stagePayload,
		keyFrom, ri.From().ServiceName(),
	)
}

func (c *profilerController) Untag(ctx context.Context) {
	c.p.Untag(ctx)
}
