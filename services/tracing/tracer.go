package tracing

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/appnet-org/arpc/pkg/rpc/element"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

// TracingElement implements RPC element interface for distributed tracing
type TracingElement struct {
	tracer opentracing.Tracer
	closer io.Closer
}

var (
	defaultSampleRatio float64 = 1
)

// NewTracingElement creates a new tracing element with Jaeger configuration
func NewTracingElement(serviceName string) (element.RPCElement, error) {
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
		return nil, fmt.Errorf("failed to initialize Jaeger tracer: %w", err)
	}
	opentracing.SetGlobalTracer(tracer)
	return &TracingElement{
		tracer: tracer,
		closer: closer,
	}, nil
}

func (t *TracingElement) Name() string {
	return "tracing"
}

func (t *TracingElement) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, error) {
	// Create span for the RPC call
	span, ctx := opentracing.StartSpanFromContext(ctx,
		fmt.Sprintf("%s.%s", req.ServiceName, req.Method),
		ext.SpanKindRPCClient,
	)

	// Add tags
	span.SetTag(string(ext.Component), "arpc-client")
	span.SetTag("rpc.id", req.ID)
	span.SetTag("rpc.service", req.ServiceName)
	span.SetTag("rpc.method", req.Method)

	// Store span in context for later use
	_ = opentracing.ContextWithSpan(ctx, span)

	// Store the context in the request for later retrieval
	// Note: You might need to modify RPCRequest to include context
	// or use a different approach to pass context through

	log.Printf("Created tracing span for request: service=%s method=%s rpcID=%d",
		req.ServiceName, req.Method, req.ID)

	return req, nil
}

func (t *TracingElement) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if resp.Error != nil {
			ext.Error.Set(span, true)
			span.SetTag("error", resp.Error.Error())
		} else {
			span.SetTag("rpc.success", true)
		}
		span.Finish()
	}

	return resp, nil
}

func (t *TracingElement) Close() error {
	if t.closer != nil {
		t.closer.Close()
	}
	return nil
}
