package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/stretchr/testify/mock"
)

// DB - мок для типа postgres.DB
type DB struct {
	mock.Mock
	Pool           *Pool
	DefaultTimeout time.Duration
}

// NewDB создает новый экземпляр мок-базы данных
func NewDB() *DB {
	return &DB{
		Pool: &Pool{},
	}
}

// GetPool возвращает пул соединений
func (db *DB) GetPool() *Pool {
	return db.Pool
}

// Pool - мок пула соединений
type Pool struct {
	mock.Mock
}

// Acquire - метод для получения соединения из пула
func (p *Pool) Acquire(ctx context.Context) (*Conn, error) {
	args := p.Called(ctx)
	return args.Get(0).(*Conn), args.Error(1)
}

// Query - метод для выполнения запроса
func (p *Pool) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	ret := p.Called(query, args)
	var rows *MockRows
	if ret.Get(0) != nil {
		rows = ret.Get(0).(*MockRows)
	}
	return rows, ret.Error(1)
}

// Conn - мок для соединения с базой данных
type Conn struct {
	mock.Mock
}

// Exec - метод для выполнения команды SQL
func (conn *Conn) Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	args = conn.Called(ctx, query, args)
	return args[0], nil
}

// Release - метод для освобождения соединения
func (conn *Conn) Release() {}

// Rows - интерфейс для работы с результатами запроса
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

// MockRows - мок результатов запроса
type MockRows struct {
	mock.Mock
	Data  []*proto.KeyData
	index int
}

// NewMockRows создает новый экземпляр MockRows
func NewMockRows() *MockRows {
	return &MockRows{}
}

// Next - проверяет наличие следующей строки
func (m *MockRows) Next() bool {
	m.index++
	return m.index <= len(m.Data)
}

// Scan - считывает значения из строки
func (m *MockRows) Scan(dest ...interface{}) error {
	if m.index == 0 || m.index > len(m.Data) {
		return errors.New("no rows")
	}

	key := m.Data[m.index-1]
	dest[0] = key.KeyID
	dest[1] = key.KeyAES
	dest[2] = key.OwnerID
	dest[3] = key.Timestamp

	return nil
}

// Err - возвращает ошибку, если она есть
func (m *MockRows) Err() error {
	if len(m.Calls) == 0 {
		return nil
	}
	return m.Called().Error(0)
}
