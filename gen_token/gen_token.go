package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var errUnsupportedSigningMethod = errors.New("unsupported signing method")

type Claims struct {
	*jwt.RegisteredClaims

	Nickname  string `json:"nickname"`
	IsTeacher bool   `json:"isTeacher"`
}

func main() {
	claimsContent, err := os.ReadFile("claims.json")
	if err != nil {
		log.Fatal(err)
	}

	claims := &Claims{
		RegisteredClaims: &jwt.RegisteredClaims{},
	}

	if err := json.Unmarshal(claimsContent, claims); err != nil {
		log.Fatal(err)
	}

	if claims.IssuedAt == nil || claims.IssuedAt.IsZero() {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
	}

	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	token, err := GenerateJWTHex(claims, jwt.SigningMethodRS256, "private_key.hex")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(token) //nolint:forbidigo
}

func GenerateJWTHex(claims jwt.Claims, signingMethod jwt.SigningMethod, privateKeyHexPath string) (string, error) {
	privateKeyHexContent, err := os.ReadFile(privateKeyHexPath)
	if err != nil {
		return "", fmt.Errorf("can't load private key: %w", err)
	}

	privateKeyContent, err := hex.DecodeString(strings.TrimSpace(string(privateKeyHexContent)))
	if err != nil {
		return "", fmt.Errorf("can't decode private key from hex: %w", err)
	}

	return GenerateJWTWithKey(claims, signingMethod, privateKeyContent)
}

func GenerateJWTWithKey(claims jwt.Claims, signingMethod jwt.SigningMethod, privateKeyContent []byte) (string, error) {
	token := jwt.NewWithClaims(signingMethod, claims)

	var (
		privateKey any
		err        error
	)

	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyContent)
	case *jwt.SigningMethodECDSA:
		privateKey, err = jwt.ParseECPrivateKeyFromPEM(privateKeyContent)
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedSigningMethod, signingMethod.Alg())
	}

	if err != nil {
		return "", fmt.Errorf("can't parse private key: %w", err)
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("can't sign token: %w", err)
	}

	return tokenString, nil
}
