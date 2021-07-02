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

	pb "github.com/HuanLiu-hotstar/proto/authority"
	"google.golang.org/grpc"
	//pb "helloworld/helloworld"

	openzipkin "github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

const (
	port = ":50055"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedAuthServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Auth(ctx context.Context, in *pb.AuthRequest) (*pb.AuthReply, error) {
	span, _ := tracer.StartSpanFromContext(ctx, "do-auth")
	defer span.Finish()
	x := rand.Intn(100) + 50
	time.Sleep(time.Duration(x) * time.Millisecond)
	log.Printf("Received: %v", in.GetData())
	return &pb.AuthReply{Message: "Hello " + in.GetData()}, nil
}

var (
	tracer *openzipkin.Tracer
)

func newTracer() *openzipkin.Tracer {
	localEndpoint, err := openzipkin.NewEndpoint("auth-grpc-server", "192.168.1.61:8082")
	if err != nil {
		log.Fatalf("Failed to create Zipkin exporter: %v", err)
	}
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")

	tracer, err = openzipkin.NewTracer(reporter, openzipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		panic(fmt.Sprintf("err:%s", err))
	}
	return tracer
}
func main() {
	newTracer()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	m := map[string]string{"name": "auth"}
	middleware := grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer, zipkingrpc.ServerTags(m)))

	s := grpc.NewServer(middleware)
	pb.RegisterAuthServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
