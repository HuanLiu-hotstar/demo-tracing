package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"

	openzipkin "github.com/openzipkin/zipkin-go"

	pb "github.com/HuanLiu-hotstar/proto/authority"
	pblt "github.com/HuanLiu-hotstar/proto/ratelimit"

	"github.com/HuanLiu-hotstar/demo-tracing/zipkin/demo/tracelib"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	// zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
)

var (
	tracer        *openzipkin.Tracer
	pcAddr        = "http://127.0.0.1:18083/pc"
	authAddr      = "http://127.0.0.1:18085/auth"
	auth2Addr     = "http://127.0.0.1:18085/auth2"
	rateAddr      = "http://127.0.0.1:18082/ratelimit"
	authGrpcAddr  = "127.0.0.1:50055"
	limitGrpcAddr = "127.0.0.1:50052"

	port       = ":18080"
	localAddr  = "192.168.1.61:8082"
	serverName = "Gateway"
)

func main() {

	f := tracelib.InitTracer(tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName))
	defer f()
	mux := http.NewServeMux()
	mux.HandleFunc("/list", list)
	mux.HandleFunc("/playback", playback)
	mux.HandleFunc("/client", client)

	h := nethttp.Middleware(opentracing.GlobalTracer(), mux)
	log.Printf("Server listening! %s ...", port)
	log.Fatal(http.ListenAndServe(port, h))

}

func playback(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracelib.StartSpanFromContext(r.Context(), r.URL.Path)
	defer span.Finish()
	req, err := getbody(r)
	if err != nil {
		Write(w, 500, fmt.Sprintf("err:%s", err))
		return
	}
	span.SetTag("ID", req.ID)
	log.Printf("req:%+v", req)
	callAuthGrpc(ctx, authGrpcAddr, "auth-client")
	callRateLimitGrpc(ctx, limitGrpcAddr, "ratelimit-client")

	//call pc
	if err := callpc(ctx, pcAddr, "pc-client", req); err != nil {
		Write(w, -103, fmt.Sprintf("err:%s", err))
		return
	}

	Write(w, 0, "success")
}

// func doclient(ctx context.Context, addr string) {
func callpc(ctx context.Context, addr, clientName string, reqData *Req) error {

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		log.Printf("err:%s", err)
		return err
	}
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req)
	defer ht.Finish()
	client := http.Client{Timeout: time.Second * 5, Transport: &nethttp.Transport{}}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("err:%s", err)
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Printf("body:%s", body)
	return nil
}

func client(w http.ResponseWriter, r *http.Request) {
	span, _ := tracelib.StartSpanFromContext(r.Context(), r.URL.Path)
	defer span.Finish()
	x := rand.Intn(10) + 3
	time.Sleep(time.Duration(x) * time.Millisecond)
	w.Write([]byte("hello client"))
}
func list(w http.ResponseWriter, r *http.Request) {
}
func getConn(address string) (*grpc.ClientConn, error) {

	// Set up a connection to the server.
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer()))}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Printf("did not connect: %v", err)
	}
	return conn, err
	// defer conn.Close()
}
func callAuthGrpc(ctx context.Context, address, name string) {
	log.Printf("addr:%s name:%s", address, name)
	span, nctx := tracelib.StartSpanFromContext(ctx, name)
	defer span.Finish()

	conn, err := getConn(address)
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(nctx, time.Second)
	defer cancel()
	r, err := c.Auth(ctx, &pb.AuthRequest{Data: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		return
	}
	log.Printf("auth: %s", r.GetMessage())
}

func callRateLimitGrpc(ctx context.Context, address, name string) {
	log.Printf("addr:%s name:%s", address, name)
	span, nctx := tracelib.StartSpanFromContext(ctx, name)
	defer span.Finish()
	conn, err := getConn(address)
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()

	c := pblt.NewRateLimitClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(nctx, time.Second)
	defer cancel()
	r, err := c.Limit(ctx, &pblt.RateLimitRequest{Data: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		return
	}
	log.Printf("auth: %s", r.GetMessage())
}
