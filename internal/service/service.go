package service

import (
	"net/http"

	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/google/uuid"
)

type HashData struct {
	Key string
}

type hashResponseWriter struct {
	http.ResponseWriter
	HashData HashData
}

type KeyData struct {
	PrivateKeyPath string
}

type TLSData struct {
	TLSCertPath string
	TLSKeyPath  string
}

type TrustedSubnet struct {
	TrustedSubnet string
}

func GenerateRandom(size int) ([]byte, error) {
	// генерируем случайную последовательность байт
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomID() string {
	return uuid.New().String() // Генерация нового UUID как keyID
}

func NewHashData(hashKey string) *HashData {
	return &HashData{
		Key: hashKey,
	}
}

func GetHashString(data []byte, key string) string {
	return base64.URLEncoding.EncodeToString(getHash(data, key))
}

func getHash(data []byte, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return h.Sum(nil)
}

func NewKeyData(privateKeyPath string) *KeyData {
	return &KeyData{
		PrivateKeyPath: privateKeyPath,
	}
}

func NewTLSData(TLSCertPath string, TLSKeyPath string) *TLSData {
	return &TLSData{
		TLSCertPath: TLSCertPath,
		TLSKeyPath:  TLSKeyPath,
	}
}
