package models

import "fmt"

var _ Secret = (*LogPass)(nil)

// LogPass пара логин пароль
type LogPass struct {
	Login    string
	Password string
}

// Type возвращает тип хранимой информации
func (c LogPass) Type() SecretType {
	return secretTypeLogPass
}

// String функция отображения приватной информации
func (c LogPass) String() string {
	return fmt.Sprintf("Login: %s, Password: %s", c.Login, c.Password)
}
