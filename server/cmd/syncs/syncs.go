package syncs

import (
	"context"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/syncrepo"
)

type database interface {
	SyncDataKeys(ctx context.Context, s syncrepo.SyncInfo) ([]*proto.KeyData, error)
	PushDataKeys(ctx context.Context, keys []*proto.KeyData) error
	SyncDataUsers(ctx context.Context, s syncrepo.SyncInfo) ([]*proto.UserData, error)

	Timeout() time.Duration
}

func NewRepo(db *postgres.DB) database {
	return syncrepo.NewSync(db)
}
