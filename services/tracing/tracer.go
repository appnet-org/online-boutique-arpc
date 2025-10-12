package tracing

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/appnet-org/arpc/pkg/metadata"
	"github.com/appnet-org/arpc/pkg/rpc/element"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

// ClientTracingElement implements RPC element interface for client-side distributed tracing
type ClientTracingElement struct {
}

// ServerTracingElement implements RPC element interface for server-side distributed tracing
type ServerTracingElement struct {
}

var (
	defaultSampleRatio float64 = 0.02 // 2% sampling
)

// NewClientTracingElement creates a new client-side tracing element
// Uses the global tracer, similar to gRPC's OpenTracingClientInterceptor
func NewClientTracingElement() element.RPCElement {
	return &ClientTracingElement{}
}

// NewServerTracingElement creates a new server-side tracing element
// Uses the global tracer, similar to gRPC's OpenTracingServerInterceptor
func NewServerTracingElement() element.RPCElement {
	return &ServerTracingElement{}
}

// ClientTracingElement methods
func (t *ClientTracingElement) Name() string {
	return "client-tracing"
}

// implements OpenTracing TextMap carrier over arpc metadata
type mdCarrier struct{ md metadata.Metadata }

// writer
func (c mdCarrier) Set(key, val string) { c.md.Set(key, val) }

// reader
func (c mdCarrier) ForeachKey(handler func(key, val string) error) error {
	if c.md == nil {
		return nil
	}
	for k, v := range c.md {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Init initializes a Jaeger tracer and returns tracer and closer, exactly like gRPC's tracing.Init
func Init(serviceName string) (opentracing.Tracer, io.Closer, error) {
	ratio := defaultSampleRatio
	log.Printf("jaeger: tracing sample ratio %f", ratio)
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "probabilistic",
			Param: ratio,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			CollectorEndpoint:   "http://jaeger:14268/api/traces",
		},
	}
	logger := jaegerlog.StdLogger
	tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(logger))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Jaeger tracer: %w", err)
	}
	return tracer, closer, nil
}

func (t *ClientTracingElement) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, context.Context, error) {
	var parentCtx opentracing.SpanContext
	if p := opentracing.SpanFromContext(ctx); p != nil {
		parentCtx = p.Context()
	}
	span := opentracing.GlobalTracer().StartSpan(
		fmt.Sprintf("%s.%s", req.ServiceName, req.Method),
		opentracing.ChildOf(parentCtx),
		ext.SpanKindRPCClient,
	)

	// gRPC parity: put span into ctx seen by downstream
	ctx = opentracing.ContextWithSpan(ctx, span)

	span.SetTag(string(ext.Component), "arpc-client")
	span.SetTag("rpc.id", req.ID)
	span.SetTag("rpc.service", req.ServiceName)
	span.SetTag("rpc.method", req.Method)

	md := metadata.FromOutgoingContext(ctx)
	if md == nil {
		md = metadata.New(map[string]string{})
	}
	_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, mdCarrier{md})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return req, ctx, nil
}

func (t *ClientTracingElement) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, context.Context, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if resp.Error != nil {
			ext.Error.Set(span, true)
			span.SetTag("error", resp.Error.Error())
		} else {
			span.SetTag("rpc.success", true)
		}
		span.Finish()
		log.Printf("Finished client tracing span for response")
	}

	return resp, ctx, nil
}

func (t *ClientTracingElement) Close() error {
	// No need to close anything, global tracer is managed separately
	return nil
}

// ServerTracingElement methods
func (t *ServerTracingElement) Name() string {
	return "server-tracing"
}

func (t *ServerTracingElement) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, context.Context, error) {
	tracer := opentracing.GlobalTracer()
	md := metadata.FromIncomingContext(ctx)

	var parentCtx opentracing.SpanContext
	if md != nil {
		if sc, err := tracer.Extract(opentracing.TextMap, mdCarrier{md}); err == nil {
			parentCtx = sc
		}
	}

	span := tracer.StartSpan(
		fmt.Sprintf("%s.%s", req.ServiceName, req.Method),
		ext.RPCServerOption(parentCtx), // gRPC-style parent handling
		ext.SpanKindRPCServer,
	)
	// gRPC parity: put span into ctx so handler/response sees it
	ctx = opentracing.ContextWithSpan(ctx, span)

	span.SetTag(string(ext.Component), "arpc-server")
	span.SetTag("rpc.id", req.ID)
	span.SetTag("rpc.service", req.ServiceName)
	span.SetTag("rpc.method", req.Method)

	return req, ctx, nil
}

func (t *ServerTracingElement) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, context.Context, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if resp.Error != nil {
			ext.Error.Set(span, true)
			span.SetTag("error", resp.Error.Error())
		} else {
			span.SetTag("rpc.success", true)
		}
		span.Finish()
		// log.Printf("Finished server tracing span for response")
	}

	return resp, ctx, nil
}

func (t *ServerTracingElement) Close() error {
	// No need to close anything, global tracer is managed separately
	return nil
}
