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
	defaultSampleRatio float64 = 1
)

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

func (t *ClientTracingElement) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, error) {
	// Create client-side span using global tracer
	span, newCtx := opentracing.StartSpanFromContext(ctx,
		fmt.Sprintf("%s.%s", req.ServiceName, req.Method),
		ext.SpanKindRPCClient,
	)

	// Add client-side tags
	span.SetTag(string(ext.Component), "arpc-client")
	span.SetTag("rpc.id", req.ID)
	span.SetTag("rpc.service", req.ServiceName)
	span.SetTag("rpc.method", req.Method)

	// Inject trace context into outgoing metadata
	md := metadata.FromOutgoingContext(newCtx)
	if md == nil {
		md = metadata.New(map[string]string{})
	}

	// Inject span context into metadata
	injectSpanContextToMetadata(span.Context(), md)
	_ = metadata.NewOutgoingContext(newCtx, md)

	log.Printf("Created client tracing span for request: service=%s method=%s rpcID=%d",
		req.ServiceName, req.Method, req.ID)

	return req, nil
}

func (t *ClientTracingElement) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, error) {
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

	return resp, nil
}

func (t *ClientTracingElement) Close() error {
	// No need to close anything, global tracer is managed separately
	return nil
}

// ServerTracingElement methods
func (t *ServerTracingElement) Name() string {
	return "server-tracing"
}

func (t *ServerTracingElement) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, error) {
	// Extract trace context from incoming metadata
	md := metadata.FromIncomingContext(ctx)
	parentSpanContext := extractSpanContextFromMetadata(md)

	// Create server-side span using global tracer
	tracer := opentracing.GlobalTracer()
	var span opentracing.Span
	if parentSpanContext != nil {
		span = tracer.StartSpan(
			fmt.Sprintf("%s.%s", req.ServiceName, req.Method),
			ext.SpanKindRPCServer,
			opentracing.ChildOf(parentSpanContext),
		)
	} else {
		span = tracer.StartSpan(
			fmt.Sprintf("%s.%s", req.ServiceName, req.Method),
			ext.SpanKindRPCServer,
		)
	}

	// Add server-side tags
	span.SetTag(string(ext.Component), "arpc-server")
	span.SetTag("rpc.id", req.ID)
	span.SetTag("rpc.service", req.ServiceName)
	span.SetTag("rpc.method", req.Method)

	log.Printf("Created server tracing span for request: service=%s method=%s rpcID=%d",
		req.ServiceName, req.Method, req.ID)

	return req, nil
}

func (t *ServerTracingElement) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if resp.Error != nil {
			ext.Error.Set(span, true)
			span.SetTag("error", resp.Error.Error())
		} else {
			span.SetTag("rpc.success", true)
		}
		span.Finish()
		log.Printf("Finished server tracing span for response")
	}

	return resp, nil
}

func (t *ServerTracingElement) Close() error {
	// No need to close anything, global tracer is managed separately
	return nil
}

// Helper functions for trace context propagation
func injectSpanContextToMetadata(spanCtx opentracing.SpanContext, md metadata.Metadata) {
	// Simple implementation - in production you'd use proper OpenTracing carriers
	// For now, we'll use a basic string representation
	if spanCtx != nil {
		// This is a simplified implementation - you'd typically use the tracer's Inject method
		md.Set("x-trace-span-context", fmt.Sprintf("%v", spanCtx))
	}
}

func extractSpanContextFromMetadata(md metadata.Metadata) opentracing.SpanContext {
	// Simple implementation - in production you'd use proper OpenTracing carriers
	spanCtxStr := md.Get("x-trace-span-context")
	if spanCtxStr == "" {
		return nil
	}

	// This is a simplified implementation - you'd typically use the tracer's Extract method
	// For now, we'll return nil and let the span be created as a root span
	// In a real implementation, you'd reconstruct the SpanContext from the metadata
	return nil
}

// Legacy support - keep the old interface for backward compatibility
type TracingElement struct {
	*ClientTracingElement
}

func NewTracingElement(serviceName string) (element.RPCElement, error) {
	log.Printf("Warning: NewTracingElement is deprecated, use NewClientTracingElement or NewServerTracingElement")
	return NewClientTracingElement(), nil
}
