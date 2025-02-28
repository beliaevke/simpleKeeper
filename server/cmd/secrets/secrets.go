package secrets

import (
	"context"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/secretsrepo"
)

type database interface {
	CreateSecret(ctx context.Context, s secretsrepo.SecretInfo) (int, error)
	GetSecret(ctx context.Context, s secretsrepo.SecretInfo) (int, error)
	GetSecretInfo(ctx context.Context, s secretsrepo.SecretInfo) (*proto.GetSecretResponse, error)
	DeleteSecret(ctx context.Context, s secretsrepo.SecretInfo) (*proto.DeleteSecretResponse, error)

	GetSecretsList(ctx context.Context, OwnerID int64) ([]*proto.SecretsList, error)
	Timeout() time.Duration
}

func NewRepo(db *postgres.DB) database {
	return secretsrepo.NewSecret(db)
}
