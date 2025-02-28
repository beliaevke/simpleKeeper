package secretsrepo

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/queries"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Secret struct {
	db *postgres.DB
}

type SecretInfo struct {
	SecretID   int    `json:"id,omitempty"`
	SecretName string `json:"name"`
	SecretType string `json:"type"`
	Content    []byte `json:"content"`
	Version    string `json:"version"`
	OwnerID    int64  `json:"ownerID,omitempty"`
	KeyID      int64  `json:"KeyID,omitempty"`
}

func NewSecret(db *postgres.DB) *Secret {
	return &Secret{
		db: db,
	}
}

func (st *Secret) Timeout() time.Duration {
	return st.db.DefaultTimeout
}

func (st *Secret) CreateSecret(ctx context.Context, s SecretInfo) (int, error) {
	tx, err := st.db.Pool.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback(ctx) //nolint
	result := st.db.Pool.QueryRow(ctx, queries.SelectSecret, s.SecretName, s.OwnerID)
	switch err := result.Scan(&s.SecretName); err {
	case pgx.ErrNoRows:
		_, err = st.db.Pool.Exec(ctx, queries.CreateSecretInsert, s.SecretName, s.SecretType, s.Content, s.OwnerID, s.KeyID)
		if err != nil {
			logger.Warnf("INSERT INTO Secrets: " + err.Error())
			return -1, err
		}
		userID, err := st.GetSecret(ctx, s)
		if err != nil {
			logger.Warnf("CreateSecret ID : " + err.Error())
			return userID, err
		}
		return userID, tx.Commit(ctx)
	case nil:
		err = errors.New("secret already exists with this name")
		if err != nil {
			logger.Warnf("INSERT INTO Secrets: " + err.Error())
			return -1, err
		}
	case err:
		logger.Warnf("Query CreateSecret: " + err.Error())
		return -1, err
	}
	return -1, tx.Commit(ctx)
}

func (st *Secret) GetSecret(ctx context.Context, s SecretInfo) (int, error) {
	if s.SecretName == "" || s.OwnerID == 0 {
		err := errors.New("name or ownerID is empty")
		if err != nil {
			logger.Warnf("GetSecret: " + err.Error())
			return -1, err
		}
	}
	result := st.db.Pool.QueryRow(ctx, queries.SelectSecret, s.SecretName, s.OwnerID)
	switch err := result.Scan(&s.SecretID); err {
	case pgx.ErrNoRows:
		return -1, nil
	case nil:
		return s.SecretID, nil
	case err:
		logger.Warnf("Query GetSecret: " + err.Error() + " ID: " + strconv.Itoa(s.SecretID) + " USER: " + strconv.FormatInt(s.OwnerID, 10))
		return -1, nil
	}
	return s.SecretID, nil
}

func (st *Secret) GetSecretsList(ctx context.Context, OwnerID int64) ([]*proto.SecretsList, error) {

	var secrets []*proto.SecretsList

	if OwnerID == 0 {
		err := errors.New("ownerID is empty")
		if err != nil {
			logger.Warnf("GetSecretsList: " + err.Error())
			return secrets, err
		}
	}
	rows, err := st.db.Pool.Query(ctx, queries.SelectSecrets, OwnerID)
	if err != nil {
		logger.Warnf("GetSecretsList: " + err.Error())
		return secrets, err
	}
	for rows.Next() {
		var secret proto.SecretsList
		if err := rows.Scan(&secret.Name, &secret.Type); err != nil {
			logger.Warnf("GetSecretsList: " + err.Error())
			return secrets, err
		}
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

func (st *Secret) GetSecretInfo(ctx context.Context, s SecretInfo) (*proto.GetSecretResponse, error) {

	secretInfo := &proto.GetSecretResponse{Name: s.SecretName}

	if s.SecretName == "" || s.OwnerID == 0 {
		secretInfo.Error = "name or ownerID is empty"
		err := errors.New(secretInfo.Error)
		if err != nil {
			logger.Warnf("GetSecretInfo: " + err.Error())
			return secretInfo, err
		}
	}

	result := st.db.Pool.QueryRow(ctx, queries.SelectSecretInfo, s.SecretName, s.OwnerID)
	switch err := result.Scan(&secretInfo.Type, &secretInfo.Content, &secretInfo.Version, &secretInfo.KeyId); err {
	case pgx.ErrNoRows:
		secretInfo.Error = "ErrNoRows"
		return secretInfo, nil
	case nil:
		secretInfo.Error = "ErrNoRows"
		return secretInfo, nil
	case err:
		secretInfo.Error = "Query GetSecretInfo: " + err.Error() + " ID: " + strconv.Itoa(s.SecretID) + " USER: " + strconv.FormatInt(s.OwnerID, 10)
		logger.Warnf(secretInfo.Error)
		return secretInfo, nil
	}

	return secretInfo, nil
}

func (st *Secret) DeleteSecret(ctx context.Context, s SecretInfo) (*proto.DeleteSecretResponse, error) {

	secretInfo := &proto.DeleteSecretResponse{}

	if s.SecretName == "" || s.OwnerID == 0 {
		secretInfo.Error = "name or ownerID is empty"
		err := errors.New(secretInfo.Error)
		if err != nil {
			logger.Warnf("DeleteSecret: " + err.Error())
			return secretInfo, err
		}
	}
	_, err := st.db.Pool.Query(ctx, queries.DeleteSecret, s.SecretName, s.OwnerID)
	if err != nil {
		secretInfo.Error = "Query DeleteSecret: " + err.Error() + " ID: " + strconv.Itoa(s.SecretID) + " USER: " + strconv.FormatInt(s.OwnerID, 10)
		logger.Warnf(secretInfo.Error)
		return secretInfo, nil
	}

	secretInfo.Success = true

	return secretInfo, nil
}
