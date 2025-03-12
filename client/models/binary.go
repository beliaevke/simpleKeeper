package models

var _ Secret = (*Binary)(nil)

// Binary произвольные бинарные данные
type Binary struct {
	Data []byte
}

// Type возвращает тип хранимой информации
func (b Binary) Type() SecretType {
	return secretTypeBinary
}

// String функция отображения приватной информации
func (b Binary) String() string {
	return "BINARY DATA"
}
