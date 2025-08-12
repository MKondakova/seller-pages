package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/golang-jwt/jwt/v5"
)

var (
	errDecodePem            = errors.New("can't decode pem")
	errKeyIsNotRsaPublicKey = errors.New("key is not RSA public key")
)

type Config struct {
	ListenPort string

	PublicKey  *rsa.PublicKey  `env:"PUBLIC_KEY,notEmpty"`
	PrivateKey *rsa.PrivateKey `env:"PRIVATE_KEY"`

	RevokedTokens []string

	ServerOpts    ServerOpts
	ProductsPath  string
	FeedbacksPath string
}

func GetConfig() (*Config, error) {
	cfg := &Config{
		ListenPort: ":8080",
		ServerOpts: ServerOpts{
			ReadTimeout:          60,
			WriteTimeout:         60,
			IdleTimeout:          60,
			MaxRequestBodySizeMb: 1,
		},
		ProductsPath:  "data/enriched_products.json",
		RevokedTokens: []string{},
	}

	opts := env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(rsa.PublicKey{}):  ParsePubKey,
			reflect.TypeOf(rsa.PrivateKey{}): ParsePrivateKey,
		},
	}

	err := env.ParseWithOptions(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("env.ParseWithOptions: %w", err)
	}

	return cfg, nil
}

type ServerOpts struct {
	ReadTimeout          int `json:"read_timeout"`
	WriteTimeout         int `json:"write_timeout"`
	IdleTimeout          int `json:"idle_timeout"`
	MaxRequestBodySizeMb int `json:"max_request_body_size_mb"`
}

// ParsePubKey public keys loader for github.com/caarlos0/env/v11 lib.
func ParsePubKey(val string) (any, error) {
	publicKey, err := hex.DecodeString(val)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString: %w", err)
	}

	pubKey, err := ParseRSAPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("keys.ParseRSAPublicKey: %w", err)
	}

	return *pubKey, nil
}

// ParsePrivateKey pkcs1 private keys loader for github.com/caarlos0/env/v11 lib.
func ParsePrivateKey(value string) (any, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString: %w", err)
	}

	block, _ := pem.Decode(decoded)
	if block == nil {
		return nil, errDecodePem
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("jwt.ParseRSAPrivateKeyFromPEM: %w", err)
	}

	return *key, nil
}

func ParseRSAPublicKey(content []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, errDecodePem
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("can't parse PKIX public key: %w", err)
	}

	public, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errKeyIsNotRsaPublicKey
	}

	return public, nil
}
