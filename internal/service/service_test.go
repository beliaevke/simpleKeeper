package service

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateRandom(t *testing.T) {
	size := 32
	randomBytes, err := GenerateRandom(size)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(randomBytes) != size {
		t.Errorf("Expected %d random bytes, got %d", size, len(randomBytes))
	}
}

func TestGenerateRandomID(t *testing.T) {
	id := GenerateRandomID()
	_, err := uuid.Parse(id)
	if err != nil {
		t.Errorf("Expected valid UUID, got error: %v", err)
	}
}

func TestNewHashData(t *testing.T) {
	key := "test-hash-key"
	hashData := NewHashData(key)
	if hashData.Key != key {
		t.Errorf("Expected Key to be %s, got %s", key, hashData.Key)
	}
}

func TestGetHashString(t *testing.T) {
	data := []byte("data to hash")
	key := "test-key"
	hashString := GetHashString(data, key)

	if hashString == "" {
		t.Errorf("Expected non-empty hash string")
	}
}

func TestGetHash(t *testing.T) {
	data := []byte("data to hash")
	key := "test-key"
	expectedHash := getHash(data, key)

	if len(expectedHash) != 32 {
		t.Errorf("Expected hash length to be 32 bytes, got %d", len(expectedHash))
	}
}
