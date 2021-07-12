package main

import (
	// "context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/HuanLiu-hotstar/demo-tracing/zipkin/demo/tracelib"
	// "github.com/opentracing-contrib/go-stdlib/nethttp"
	// "github.com/opentracing/opentracing-go"
)

var (
	// tracer     *openzipkin.Tracer
	serverName = "UM-Server"
	localAddr  = "192.168.1.61"
	zipkinAddr = "http://zipkin:9411/api/v2/spans" //http://localhost:9411/api/v2/span
	port       = 18084
)

func main() {

	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),
		tracelib.WithZipkinSerAddr(zipkinAddr),
	}
	f := tracelib.InitTracer(opts...)
	defer f()

	mux := http.NewServeMux()
	mux.HandleFunc("/um", um)

	//h := nethttp.Middleware(opentracing.GlobalTracer(), mux)
	h := tracelib.MiddlewareHttp(mux)

	log.Printf("Server listening! %d ...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), h))

}
func um(w http.ResponseWriter, r *http.Request) {
	span, _ := tracelib.StartSpanFromContext(r.Context(), r.URL.Path)
	defer span.Finish()
	x := rand.Intn(100) + 10
	time.Sleep(time.Duration(x) * time.Millisecond)
	w.Write([]byte(fmt.Sprintf(`{"um":%d}`, x)))
	log.Printf("res:%d", x)
}
