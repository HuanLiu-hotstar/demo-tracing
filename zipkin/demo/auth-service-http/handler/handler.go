package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/dizys/ambassador-kustomization-example/auth-service-http/config"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	*jwt.StandardClaims
	Id       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

type Handler struct {
}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	authStr := req.Header.Get("Authorization")

	if config.Config.GetBool("request_logging") {
		log.Printf("[Request] %s - %s (token: %s): %s\n", req.Method, req.RequestURI, authStr, req.PostForm.Encode())
	}

	if authStr == "" {
		rErr(resp, 401, "Unauthenticated")
		return
	}

	if !strings.HasPrefix(authStr, "Bearer ") {
		rErr(resp, 401, "Invalid access token type")
		return
	}

	unverifiedToken := strings.TrimPrefix(authStr, "Bearer ")

	pubKeyPEM := config.Config.GetString("jwt_rsa_public_key")

	pubKey, err := PEMStringToRSAPublicKey(pubKeyPEM)

	if err != nil {
		rErr(resp, 503, "Invalid public key")
		return
	}

	token, err := jwt.ParseWithClaims(unverifiedToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return pubKey, nil
	})

	if err != nil {
		rErr(resp, 401, "Unauthorized")
		return
	}

	claims := token.Claims.(*Claims)

	claimsStr, err := StructToJSON(claims)

	if err != nil {
		rErr(resp, 503, "Cannot convert claims to JSON")
		return
	}

	resp.Header().Add("x-passport", claimsStr)

	resp.Write([]byte("OK"))

	if config.Config.GetBool("request_logging") {
		log.Printf("[Response] 200: OK\n")
	}
}
