# usage for tracinglib

## http server

- http server use middleware [demo here](https://github.com/HuanLiu-hotstar/demo-tracing/blob/main/zipkin/demo/um/main.go#L24)

```go
	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),tracelib.WithZipkinSerAddr(zipkinAddr)}
	f := tracelib.InitTracer(opts...)
	defer f() // release resources
	mux := http.NewServeMux()
	mux.HandleFunc("/pc", pc)
	//middleware for extract http header to restore context 
	h := tracelib.MiddlewareHttp(mux) 
	log.Printf("Server listening! %d ...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), h))

```

## http client

- http [demo here](https://github.com/HuanLiu-hotstar/demo-tracing/blob/main/zipkin/demo/pc/main.go#L63)


```go
opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),tracelib.WithZipkinSerAddr(zipkinAddr)}
	f := tracelib.InitTracer(opts...)
	defer f() // release resources
httpcilent:= tracelib.NewHttpRequest(tracelib.WithTimeout(5*time.Second))
respbyte,err := httpclient.Do(ctx,body)


```

## gin server

- gin server use middleware [demo here](https://github.com/HuanLiu-hotstar/demo-tracing/blob/main/zipkin/demo/pc/main.go#L35)

```go
	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),tracelib.WithZipkinSerAddr(zipkinAddr)}

	f := tracelib.InitTracer(opts...)
	defer f() // release resources
	rgin := gin.Default()
	
	rgin.Use(tracelib.MiddlewareGin())

	handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 200, "msg": "OK1"})
			}
	port := ":8080"
	rgin.GET("/list",handler)
	rgin.Run(port)
	log.Printf("Server listening! %s ...", port)
```

## grpc server 

- grpc server use intercept to handler tracing data
- [demo here](https://github.com/HuanLiu-hotstar/demo-tracing/blob/main/zipkin/demo/grpc/ratelimit_server/main.go#L103) 

```go 
	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),
		tracelib.WithZipkinSerAddr(zipkinAddr),
	}
	f := tracelib.InitTracer(opts...)
	defer f()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Initialize the gRPC server.
	s := tracelib.NewGrpcServer()
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
```

## grpc client 

- nearly same as grpc server
- [demo here](https://github.com/HuanLiu-hotstar/demo-tracing/blob/main/zipkin/demo/gateway/main.go#L164)

```go
	opts := []tracelib.ConfigOpt{tracelib.WithLocalAddr(localAddr), tracelib.WithLocalName(serverName),
		tracelib.WithZipkinSerAddr(zipkinAddr),
	}
	f := tracelib.InitTracer(opts...)
	defer f()
	conn, err := tracelib.NewGrpcConn(address)
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthClient(conn)
	nctx := context.Backgroud()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(nctx, time.Second)
	defer cancel()
	r, err := c.Auth(ctx, &pb.AuthRequest{Data: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		return
	}
	log.Printf("auth: %s", r.GetMessage())
```

## span usage

```go
	// create a sub-span ctx is a context from other span ,
	// then this will generate a sub-span from parent span 
	span ,subcxt := tracelib.StartSpanFromContext(ctx,"sub-span")
	defer span.Finish()
	// use ctx for other 
```

## tag usage

```go

	span ,_ := tracelib.StartSpanFromContext(ctx,"sub-span")
	defer span.Finish()
	// we can set tag for business id,that will be used for debug 
	span.Tag("ID",ID) 
```

## other usage

- reference this docs about [opentracing tracing ](https://github.com/opentracing/opentracing-go)