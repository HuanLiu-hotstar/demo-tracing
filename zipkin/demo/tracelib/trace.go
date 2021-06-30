package tracelib

import (
	"context"
	"fmt"
	// "log"
	// "math/rand"
	// "net/http"
	// "strings"
	// "time"

	"github.com/opentracing/opentracing-go"
	openzipkin "github.com/openzipkin/zipkin-go"
	//zipkinmw "github.com/openzipkin/zipkin-go/middleware/http"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

var (
	tracer            *openzipkin.Tracer
	zipkinDefaultAddr = "http://localhost:9411/api/v2/spans"
)

type Config struct {
	LocalAddr     string
	LocalName     string
	ZipkinSerAddr string
}
type ConfigOpt func(*Config)

func WithLocalAddr(localName string) ConfigOpt {
	return func(c *Config) {
		c.LocalAddr = localName
	}
}
func WithLocalName(localName string) ConfigOpt {
	return func(c *Config) {
		c.LocalName = localName
	}
}
func InitTracer(opts ...ConfigOpt) func() {
	c := Config{
		LocalAddr:     "localhost",
		LocalName:     "localhost",
		ZipkinSerAddr: zipkinDefaultAddr,
	}
	for _, o := range opts {
		o(&c)
	}

	localEndpoint, err := openzipkin.NewEndpoint(c.LocalName, c.LocalAddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Zipkin exporter: %v", err))
	}
	reporter := zipkinHTTP.NewReporter(c.ZipkinSerAddr)
	f := func() { reporter.Close() }

	nativeTracer, err := openzipkin.NewTracer(reporter, openzipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		panic(fmt.Sprintf("err:%s", err))
	}
	tracer = nativeTracer
	opentracing.SetGlobalTracer(zipkinot.Wrap(nativeTracer))
	return f
}
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, operationName, opts...)
}

func StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return opentracing.StartSpan(operationName, opts...)
}
