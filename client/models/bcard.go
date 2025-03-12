package models

import (
	"fmt"
)

var _ Secret = (*BCard)(nil)

// BCard данные банковской карты
type BCard struct {
	Number       string
	Holder       string
	ExpiryDate   string
	SecurityCode string
}

// Type возвращает тип хранимой информации
func (c BCard) Type() SecretType {
	return secretTypeBCard
}

// String функция отображения приватной информации
func (c BCard) String() string {
	return fmt.Sprintf("Number: %s, Holder: %s, ExpiryDate: %s, SecurityCode: %s",
		c.Number, c.Holder, c.ExpiryDate, c.SecurityCode)
}
