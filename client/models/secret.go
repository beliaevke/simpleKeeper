package models

import (
	"encoding/json"
	"errors"
)

// SecretType тип секрета
type SecretType string

const (
	secretTypeLogPass SecretType = "logpass"
	secretTypeText    SecretType = "text"
	secretTypeBinary  SecretType = "binary"
	secretTypeBCard   SecretType = "bcard"
)

// Secret приватные данные пользователя
type Secret interface {
	// Type возвращает тип хранимой информации
	Type() SecretType
	// String функция отображения приватной информации
	String() string
}

type container struct {
	Type SecretType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

// DecodeSecret функция декодирования данных пользователя
func DecodeSecret(data string) (Secret, error) {
	var c container
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		return nil, err
	}

	switch c.Type {
	case secretTypeLogPass:
		var logpass LogPass
		if err := json.Unmarshal(c.Data, &logpass); err != nil {
			return nil, err
		}
		return logpass, nil
	case secretTypeText:
		var text Text
		if err := json.Unmarshal(c.Data, &text); err != nil {
			return nil, err
		}
		return text, nil
	case secretTypeBinary:
		var binary Binary
		if err := json.Unmarshal(c.Data, &binary); err != nil {
			return nil, err
		}
		return binary, nil
	case secretTypeBCard:
		var bcard BCard
		if err := json.Unmarshal(c.Data, &bcard); err != nil {
			return nil, err
		}
		return bcard, nil
	default:
		return nil, errors.New("unknown secret type")
	}
}

// EncodeSecret функция кодирования данных пользователя
func EncodeSecret(secret Secret) (string, error) {
	data, err := json.Marshal(secret)
	if err != nil {
		return "", err
	}
	secdata, err := json.Marshal(container{
		Type: secret.Type(),
		Data: data,
	})
	return string(secdata), err
}
