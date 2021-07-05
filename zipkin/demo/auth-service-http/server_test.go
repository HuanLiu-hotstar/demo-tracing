package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/dizys/ambassador-kustomization-example/auth-service-http/config"
	"github.com/dizys/ambassador-kustomization-example/auth-service-http/handler"
	"github.com/golang-jwt/jwt"
	"gotest.tools/assert"
)

func init() {
	config.SetupConfig()
}

func TestGenerateKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)

	assert.NilError(t, err)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)

	assert.NilError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())

	assert.NilError(t, err)

	privateKeyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	t.Logf("private:\n%s\n", string(privateKeyPEMBytes))
	t.Logf("public:\n%s\n", string(publicKeyPEMBytes))
}

func TestGenerateTokens(t *testing.T) {
	priKeyPEM := config.Config.GetString("jwt_rsa_private_key")

	priKey, err := handler.PEMStringToRSAPrivateKey(priKeyPEM)

	assert.NilError(t, err)

	claims1 := &handler.Claims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 3).Unix(),
		},
		Id:       1,
		Username: "user1",
	}

	claims2 := &handler.Claims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 3).Unix(),
		},
		Id:       2,
		Username: "user2",
	}

	claims3 := &handler.Claims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 3).Unix(),
		},
		Id:       3,
		Username: "user3",
	}

	assert.NilError(t, err)

	token1, err := generateToken(priKey, claims1)

	assert.NilError(t, err)

	t.Logf("token1:\n%s\n", token1)

	token2, err := generateToken(priKey, claims2)

	assert.NilError(t, err)

	t.Logf("token2:\n%s\n", token2)

	token3, err := generateToken(priKey, claims3)

	assert.NilError(t, err)

	t.Logf("token3:\n%s\n", token3)
}

func generateToken(priKey *rsa.PrivateKey, claims *handler.Claims) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("RS256"))

	token.Claims = claims

	return token.SignedString(priKey)
}
