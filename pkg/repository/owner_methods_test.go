package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// helper to create repository and db
func setupOwnerRepo(t *testing.T, enforce bool, defaultID interface{}) (*OwnerGenericRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test",
		Model: &OwnerTestEntity{},
	})

	ownerRes := resource.NewOwnerResource(res, resource.OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: enforce,
		DefaultOwnerID:   defaultID,
	})

	repoIface, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)
	return repoIface.(*OwnerGenericRepository), db
}

func TestExtractOwnerID(t *testing.T) {
	repo, _ := setupOwnerRepo(t, true, nil)
	ctx := context.WithValue(context.Background(), middleware.OwnerContextKey, "owner-a")

	id, err := repo.extractOwnerID(ctx)
	require.NoError(t, err)
	assert.Equal(t, "owner-a", id)

	_, err = repo.extractOwnerID(context.Background())
	assert.Equal(t, ErrOwnerIDNotFound, err)

	repoDefault, _ := setupOwnerRepo(t, true, "def")
	id, err = repoDefault.extractOwnerID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "def", id)

	repoDisabled, _ := setupOwnerRepo(t, false, "zzz")
	id, err = repoDisabled.extractOwnerID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "zzz", id)
}

func TestApplyOwnerFilter(t *testing.T) {
	repo, db := setupOwnerRepo(t, true, nil)
	ctx := context.WithValue(context.Background(), middleware.OwnerContextKey, "owner-a")

	tx, err := repo.applyOwnerFilter(ctx, db.Session(&gorm.Session{DryRun: true}))
	require.NoError(t, err)
	tx.Find(&[]OwnerTestEntity{})
	sql := tx.Statement.SQL.String()
	assert.Contains(t, sql, "owner_id")
	assert.Equal(t, "owner-a", tx.Statement.Vars[len(tx.Statement.Vars)-1])

	_, err = repo.applyOwnerFilter(context.Background(), db)
	assert.Equal(t, ErrOwnerIDNotFound, err)

	repoDisabled, _ := setupOwnerRepo(t, false, nil)
	tx, err = repoDisabled.applyOwnerFilter(ctx, db.Session(&gorm.Session{DryRun: true}))
	require.NoError(t, err)
	tx.Find(&[]OwnerTestEntity{})
	sql = tx.Statement.SQL.String()
	assert.NotContains(t, sql, "owner_id")
}

func TestVerifyOwnership(t *testing.T) {
	repo, db := setupOwnerRepo(t, true, nil)
	items := []OwnerTestEntity{{Name: "A1", OwnerID: "owner-a"}, {Name: "B1", OwnerID: "owner-b"}}
	require.NoError(t, db.Create(&items).Error)

	ctxA := context.WithValue(context.Background(), middleware.OwnerContextKey, "owner-a")
	err := repo.verifyOwnership(ctxA, items[0].ID)
	require.NoError(t, err)

	err = repo.verifyOwnership(ctxA, items[1].ID)
	assert.Equal(t, ErrOwnerMismatch, err)

	err = repo.verifyOwnership(ctxA, uint(9999))
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	err = repo.verifyOwnership(context.Background(), items[0].ID)
	assert.Equal(t, ErrOwnerIDNotFound, err)
}

func TestSetOwnership(t *testing.T) {
	repo, _ := setupOwnerRepo(t, true, nil)
	ctx := context.WithValue(context.Background(), middleware.OwnerContextKey, "owner-a")

	e := &OwnerTestEntity{Name: "one"}
	require.NoError(t, repo.setOwnership(ctx, e))
	assert.Equal(t, "owner-a", e.OwnerID)

	list := []OwnerTestEntity{{Name: "l1"}, {Name: "l2"}}
	require.NoError(t, repo.setOwnership(ctx, &list))
	for _, it := range list {
		assert.Equal(t, "owner-a", it.OwnerID)
	}

	type Bad struct{ Name string }
	err := repo.setOwnership(ctx, &Bad{Name: "x"})
	assert.Error(t, err)

	repoDisabled, _ := setupOwnerRepo(t, false, nil)
	item := &OwnerTestEntity{Name: "n"}
	require.NoError(t, repoDisabled.setOwnership(ctx, item))
	assert.Equal(t, "", item.OwnerID)
}
