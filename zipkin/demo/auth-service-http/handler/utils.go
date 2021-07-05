package handler

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"

	"github.com/dizys/ambassador-kustomization-example/auth-service-http/config"
)

func rErr(resp http.ResponseWriter, statusCode int, message string) {
	resp.WriteHeader(statusCode)
	resp.Write([]byte(message))

	if config.Config.GetBool("request_logging") {
		log.Printf("[Response] %d: %s\n", statusCode, message)
	}
}

func PEMStringToRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	p, _ := pem.Decode([]byte(pemStr))

	pub, err := x509.ParsePKIXPublicKey(p.Bytes)

	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}

	return nil, fmt.Errorf("public key type is not RSA")
}

func PEMStringToRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	p, _ := pem.Decode([]byte(pemStr))

	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func StructToJSON(obj interface{}) (string, error) {
	jsonBytes, err := json.Marshal(obj)

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
