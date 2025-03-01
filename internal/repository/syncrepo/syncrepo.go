package syncrepo

import (
	"context"
	"errors"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/queries"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Sync struct {
	db *postgres.DB
}

type SyncInfo struct {
	OwnerID int64 `json:"ownerID,omitempty"`
}

func NewSync(db *postgres.DB) *Sync {
	return &Sync{
		db: db,
	}
}

func (sc *Sync) Timeout() time.Duration {
	return sc.db.DefaultTimeout
}

func (sc *Sync) SyncDataKeys(ctx context.Context, si SyncInfo) ([]*proto.KeyData, error) {

	var keys []*proto.KeyData

	if si.OwnerID == 0 {
		err := errors.New("ownerID is empty")
		if err != nil {
			logger.Warnf("SyncDataKeys: " + err.Error())
			return keys, err
		}
	}
	rows, err := sc.db.Pool.Query(ctx, queries.SyncDataKeys, si.OwnerID)
	if err != nil {
		logger.Warnf("SyncDataKeys: " + err.Error())
		return keys, err
	}
	for rows.Next() {
		var key proto.KeyData
		if err := rows.Scan(&key.KeyID, &key.KeyAES, &key.OwnerID, &key.Timestamp); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	return keys, nil
}

func (sc *Sync) PushDataKeys(ctx context.Context, keys []*proto.KeyData) error {

	conn, err := sc.db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	for _, key := range keys {
		_, err = conn.Exec(ctx, queries.PushDataKeys, key.KeyID, key.KeyAES, key.OwnerID, key.Timestamp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Sync) SyncDataUsers(ctx context.Context, si SyncInfo) ([]*proto.UserData, error) {

	var users []*proto.UserData

	if si.OwnerID == 0 {
		err := errors.New("ownerID is empty")
		if err != nil {
			logger.Warnf("SyncDataUsers: " + err.Error())
			return users, err
		}
	}
	rows, err := sc.db.Pool.Query(ctx, queries.SyncDataUsers, si.OwnerID)
	if err != nil {
		logger.Warnf("SyncDataUsers: " + err.Error())
		return users, err
	}
	for rows.Next() {
		var user proto.UserData
		if err := rows.Scan(&user.UserID, &user.UserLogin, &user.UserPassword); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (sc *Sync) SyncDataSecrets(ctx context.Context, si SyncInfo) ([]*proto.SecretData, error) {

	var secrets []*proto.SecretData

	if si.OwnerID == 0 {
		err := errors.New("ownerID is empty")
		if err != nil {
			logger.Warnf("SyncDataSecrets: " + err.Error())
			return secrets, err
		}
	}
	rows, err := sc.db.Pool.Query(ctx, queries.SyncDataSecrets, si.OwnerID)
	if err != nil {
		logger.Warnf("SyncDataSecrets: " + err.Error())
		return secrets, err
	}
	for rows.Next() {
		var secret proto.SecretData
		if err := rows.Scan(&secret.Name, &secret.Type, &secret.Content, &secret.OwnerID, &secret.KeyID, &secret.Timestamp, &secret.IsDeleted); err != nil {
			return nil, err
		}
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

func (sc *Sync) PushDataSecrets(ctx context.Context, secrets []*proto.SecretData) error {

	conn, err := sc.db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	for _, secret := range secrets {
		_, err = conn.Exec(ctx, queries.PushDataSecrets, secret.Name, secret.Type, secret.Content, secret.OwnerID, secret.KeyID, secret.Timestamp, secret.IsDeleted)
		if err != nil {
			return err
		}
	}
	return nil
}
