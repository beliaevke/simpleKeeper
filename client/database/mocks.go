package database

import (
	"database/sql"

	"github.com/beliaevke/simpleKeeper/internal/proto"

	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
	Secrets   map[string]*proto.GetSecretResponse
	UserID    int
	LoginErr  error
	SecretErr error
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.SecretErr != nil {
		return nil, m.SecretErr
	}
	return nil, nil
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ret := m.Called(query, args)
	var r *sql.Rows
	if ret.Get(0) != nil {
		r = ret.Get(0).(*sql.Rows)
	}
	return r, ret.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.LoginErr != nil {
		return &sql.Row{}
	}
	return &sql.Row{}
}

func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	return nil, nil
}

func (m *MockDB) Close() error {
	return nil
}
