package syncrepo

import (
	"context"
	"testing"
	"time"

	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/db/postgres/mocks"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/queries"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stretchr/testify/assert"
)

func newSync(db *mocks.DB) *Sync {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, "postgres://postgres:pos111@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		return nil
	}
	pos := &postgres.DB{Pool: dbpool, DefaultTimeout: time.Second * 5}
	return &Sync{
		db: pos,
	}
}

func TestNewSync(t *testing.T) {
	db := new(mocks.DB)
	sync := newSync(db)

	assert.NotNil(t, sync)
	assert.Equal(t, db, sync.db)
}

func TestSync_Timeout(t *testing.T) {
	db := &mocks.DB{DefaultTimeout: 10 * time.Second}
	sync := newSync(db)

	assert.Equal(t, 10*time.Second, sync.Timeout())
}

func TestSyncDataKeys_EmptyOwnerID(t *testing.T) {
	db := new(mocks.DB)
	sync := newSync(db)

	keys, err := sync.SyncDataKeys(context.Background(), SyncInfo{OwnerID: 0})

	assert.Empty(t, keys)
	assert.EqualError(t, err, "ownerID is empty")
}

func TestSyncDataKeys_QueryError(t *testing.T) {
	mockDB := mocks.NewDB()
	sync := newSync(mockDB)

	rows := &mocks.MockRows{
		Data: []*proto.KeyData{
			{KeyID: "key1", KeyAES: "aes1", OwnerID: 1, Timestamp: time.Now().UTC().String()},
		},
	}
	args := []interface{}{int64(1)}

	mockDB.Pool.On("Query", queries.SyncDataKeys, args).
		Return(rows, nil)

	keys, err := sync.SyncDataKeys(context.Background(), SyncInfo{OwnerID: 1})

	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	mockDB.AssertExpectations(t)
}
