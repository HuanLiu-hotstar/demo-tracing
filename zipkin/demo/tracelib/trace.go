package tracelib

import (
	"bytes"
	"context"
	"fmt"
	// "log"
	// "math/rand"
	"io/ioutil"
	"net/http"
	// "strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

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
func WithZipkinSerAddr(addr string) ConfigOpt {
	return func(c *Config) {
		c.ZipkinSerAddr = addr
	}
}

// InitTracer
// usage:
// df := InitTracer(opts...)
// defer df()
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

//StartSpanFromContext start a sub-span from context
// usage:
// span,pctx := tracelib.StartSpanFromContext(ctx,"span_name")
// defer span.Finish()
// use pctx for generating other span

func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, operationName, opts...)
}

// StartSpan start a new span
func StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return opentracing.StartSpan(operationName, opts...)
}

// NewGrpcConn new client connection to grpc server
// usage:
// conn,err := NewGrpcConn(address,opts...)
// defer conn.Close()
func NewGrpcConn(address string, opt ...grpc.DialOption) (*grpc.ClientConn, error) {

	// Set up a connection to the server.
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	}
	opts = append(opts, opt...)
	conn, err := grpc.Dial(address, opts...)
	return conn, err
}

// NewGrpcServer new a grpc server with tracing
// usage:
// s := NewGrpcConn(opts...)
// err := s.Serve()
func NewGrpcServer(opt ...grpc.ServerOption) *grpc.Server {

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
		grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer())),
	}
	opts = append(opts, opt...)
	s := grpc.NewServer(opts...)
	return s
}

// MiddlewareHttp generate middleware for http.Handler
// usage:
// h:= MiddlewareHttp(my http.Handler)
// http.ListenAndServe(port,h)
func MiddlewareHttp(h http.Handler, options ...nethttp.MWOption) http.Handler {
	return nethttp.Middleware(opentracing.GlobalTracer(), h, options...)
}

// MiddlewareGin generate the middleware of gin
// usage:
// g.Default().Use(MiddlewareGin())
func MiddlewareGin(options ...nethttp.MWOption) func(c *gin.Context) {
	return ginhttp.Middleware(opentracing.GlobalTracer())
}

type HttpRequest struct {
	Method  string
	Address string
	Data    []byte
	Timeout time.Duration
	header  map[string]string
}

func NewHttpRequest() HttpRequest {
	return HttpRequest{
		Method:  "GET",
		Timeout: time.Second * 5,
		header:  make(map[string]string),
	}
}
func (h HttpRequest) Do(ctx context.Context, b []byte) ([]byte, error) {
	return NewHttpReq(ctx, h.Method, h.Address, b, h.Timeout)
}
func (h HttpRequest) AddHeader(m map[string]string) {
	// h.header = m
	for k, v := range m {
		h.header[k] = v
	}
}

// NewHttpReq with context and with a span
func NewHttpReq(ctx context.Context, method, addr string, bye []byte, timeout time.Duration) ([]byte, error) {

	req, err := http.NewRequest(method, addr, bytes.NewBuffer(bye))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req)
	defer ht.Finish()
	client := http.Client{Timeout: timeout, Transport: &nethttp.Transport{}}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err
}
