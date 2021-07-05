/*
 *
 * Copyright 2015 gRPC authors.
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
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/HuanLiu-hotstar/demo-tracing/zipkin/demo/tracelib"
	pb "github.com/HuanLiu-hotstar/proto/ratelimit"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	//pb "helloworld/helloworld"

	openzipkin "github.com/openzipkin/zipkin-go"
	// zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

const (
	port       = ":50052"
	localAddr  = "192.168.1.61:8082"
	serverName = "ratelimit"
	zipkinAddr = "http://zipkin:9411/api/v2/spans" //http://localhost:9411/api/v2/span
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedRateLimitServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Limit(ctx context.Context, in *pb.RateLimitRequest) (*pb.RateLimitReply, error) {
	span, _ := tracer.StartSpanFromContext(ctx, "do-ratelimit")
	defer span.Finish()
	x := rand.Intn(50) + 50
	time.Sleep(time.Duration(x) * time.Millisecond)
	log.Printf("Received: %v", in.GetData())
	return &pb.RateLimitReply{Message: "rateLimit " + in.GetData()}, nil
}

var (
	tracer *openzipkin.Tracer
)

func newTracer() *openzipkin.Tracer {
	localEndpoint, err := openzipkin.NewEndpoint("limit-grpc-server", "192.168.1.61:8082")
	if err != nil {
		log.Fatalf("Failed to create Zipkin exporter: %v", err)
	}
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")
	// defer reporter.Close()

	tracer, err = openzipkin.NewTracer(reporter, openzipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		panic(fmt.Sprintf("err:%s", err))
	}
	return tracer
}
func main() {
	// newTracer()
	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),
		tracelib.WithZipkinSerAddr(zipkinAddr),
	}
	f := tracelib.InitTracer(opts...)
	defer f()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// m := map[string]string{"name": "ratelimit"}
	// middleware := grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer, zipkingrpc.ServerTags(m)))
	// s := grpc.NewServer(middleware)
	// Initialize the gRPC server.
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
		grpc.StreamInterceptor(
			otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer())))
	pb.RegisterRateLimitServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
