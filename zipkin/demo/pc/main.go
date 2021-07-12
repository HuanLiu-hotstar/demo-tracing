package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HuanLiu-hotstar/demo-tracing/zipkin/demo/tracelib"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

var (
	serverName = "PC-Server"
	localAddr  = "192.168.1.63"
	port       = 18083
	umAddr     = "http://um-service-http:18084/um"
	zipkinAddr = "http://zipkin:9411/api/v2/spans" //http://localhost:9411/api/v2/span
)

func main() {

	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),
		tracelib.WithZipkinSerAddr(zipkinAddr),
	}
	f := tracelib.InitTracer(opts...)
	defer f()
	rgin := gin.Default()
	rgin.Use(tracelib.MiddlewareGin())
	rgin.GET("/pc", func(c *gin.Context) {
		pc(c.Writer, c.Request)
	})
	rgin.Run(fmt.Sprintf(":%d", port))

	// h := nethttp.Middleware(opentracing.GlobalTracer(), mux)
	//	h := tracelib.MiddlewareHttp(mux)
	//	log.Printf("Server listening! %d ...", port)
	//	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), h))

}
func pc(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracelib.StartSpanFromContext(r.Context(), r.URL.Path)
	defer span.Finish()
	doclient(ctx)
	x := rand.Intn(100) + 10
	otherwork(ctx)

	w.Write([]byte(fmt.Sprintf(`{"pc":%d}`, x)))
}
func otherwork(c context.Context) {
	span, _ := tracelib.StartSpanFromContext(c, "other-work")
	defer span.Finish()
	x := rand.Intn(100) + 10
	time.Sleep(time.Duration(x) * time.Millisecond)
}

func doclient(ctx context.Context) {
	opts := []tracelib.HttpRequestOpt{
		tracelib.WithAddress(umAddr),
	}
	httpclient := tracelib.NewHttpRequest(opts...)
	respbyte, err := httpclient.Do(ctx, nil)
	if err != nil {
		log.Printf("err:%s", err)
		return
	}
	log.Printf("body:%s", respbyte)
}
func doclient2(ctx context.Context) {

	req, err := http.NewRequest("GET", umAddr, nil)
	if err != nil {
		log.Printf("err:%s", err)
		return
	}
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req)
	defer ht.Finish()
	client := http.Client{Timeout: time.Second * 5, Transport: &nethttp.Transport{}}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("err:%s", err)
		return
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Printf("body:%s", body)

}
