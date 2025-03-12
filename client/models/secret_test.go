package models

import (
	"encoding/json"
	"testing"
)

func TestEncodeDecodeSecret(t *testing.T) {
	tests := []struct {
		secret   Secret
		expected string
	}{
		{
			secret: LogPass{
				Login:    "user1",
				Password: "pass1",
			},
			expected: `{"type":"logpass","data":{"Login":"user1","Password":"pass1"}}`,
		},
		{
			secret: Text{
				Data: "This is a secret text",
			},
			expected: `{"type":"text","data":{"Data":"This is a secret text"}}`,
		},
	}

	for _, test := range tests {
		encoded, err := EncodeSecret(test.secret)
		if err != nil {
			t.Errorf("failed to encode secret: %v", err)
			continue
		}

		var got string = encoded
		var expected string = test.expected
		if got != expected {
			t.Errorf("expected encoded secret to be %q, got %q", expected, got)
		}

		decoded, err := DecodeSecret(encoded)
		if err != nil {
			t.Errorf("failed to decode secret: %v", err)
			continue
		}

		if decoded.Type() != test.secret.Type() {
			t.Errorf("expected secret type %q, got %q", test.secret.Type(), decoded.Type())
		}

		if !jsonEqual(test.secret, decoded) {
			t.Errorf("decoded secret does not match the original")
		}
	}
}

func jsonEqual(a, b Secret) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

func TestDecodeSecretInvalidJSON(t *testing.T) {
	_, err := DecodeSecret("invalid json")
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}

func TestDecodeSecretUnknownType(t *testing.T) {
	data := `{"type":"unknown","data":{}}`
	_, err := DecodeSecret(data)
	if err == nil {
		t.Errorf("expected error for unknown secret type, got nil")
	}
}
