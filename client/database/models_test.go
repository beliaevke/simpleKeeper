package database

import (
	"testing"

	"github.com/beliaevke/simpleKeeper/internal/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAddSecret(t *testing.T) {
	mockDB := &MockDB{}
	secretRequest := &proto.CreateSecretRequest{
		Name:    "testSecret",
		Type:    "type1",
		Content: []byte("content1"),
		UserId:  1,
		KeyId:   1,
	}

	err := AddSecret(mockDB, secretRequest)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestDeleteSecret(t *testing.T) {
	mockDB := &MockDB{}
	deleteRequest := &proto.DeleteSecretRequest{
		Name:   "testSecret",
		UserId: 1,
	}

	err := DeleteSecret(mockDB, deleteRequest)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestLoginUser_EmptyCredentials(t *testing.T) {
	mockDB := &MockDB{}

	userID, err := LoginUser(mockDB, "", "")
	if err == nil || userID != -1 {
		t.Errorf("Expected error for empty username and password, got userID %v, err: %v", userID, err)
	}
}

func TestGetSecretsList(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection\n", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name", "type"}).
		AddRow("secret1", "type1").
		AddRow("secret2", "type2")

	mock.ExpectQuery("SELECT name, type FROM Secrets WHERE ownerID = ?").
		WithArgs(1).
		WillReturnRows(rows)

	secrets, err := GetSecretsList(db, 1)

	assert.NoError(t, err)

	assert.Len(t, secrets, 2)
	assert.Equal(t, "secret1", secrets[0].Name)
	assert.Equal(t, "type1", secrets[0].Type)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetSecret(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection\n", err)
	}
	defer db.Close()

	secretName := "my_secret"
	userID := 1
	secretType := "TEXT"
	content := []byte("s3cr3t_content")
	version := 1
	keyID := int64(123)

	mock.ExpectQuery("SELECT type, content, version, keyID FROM Secrets WHERE name = \\? AND ownerID = \\?").
		WithArgs(secretName, userID).
		WillReturnRows(sqlmock.NewRows([]string{"type", "content", "version", "keyID"}).
			AddRow(secretType, content, version, keyID))

	secret, err := GetSecret(db, secretName, userID)

	assert.NoError(t, err)

	assert.Equal(t, secretType, secret.Type)
	assert.Equal(t, keyID, secret.KeyId)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetSecretWrongUserID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection\n", err)
	}
	defer db.Close()

	secret, err := GetSecret(db, "my_secret", -1)

	assert.Error(t, err)
	assert.Equal(t, "wrong userID", err.Error())
	assert.Equal(t, &proto.GetSecretResponse{}, secret)
}
