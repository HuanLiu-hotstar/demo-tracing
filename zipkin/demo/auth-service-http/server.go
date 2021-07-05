package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dizys/ambassador-kustomization-example/auth-service-http/config"
	"github.com/dizys/ambassador-kustomization-example/auth-service-http/handler"
)

func main() {
	config.SetupConfig()

	port := config.Config.GetInt("port")

	if port <= 0 || port >= 65_535 {
		log.Fatalf("Invalid port number: %d\n", port)
		os.Exit(1)
	}

	log.Printf("Auth service (HTTP) running on %d...\n", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), &handler.Handler{})
	if err != nil {
		panic(err)
	}
}
