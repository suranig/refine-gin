package repository

import (
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type idModel struct{ UID string }

func TestGetIDFieldName(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	res := resource.NewResource(resource.ResourceConfig{Model: &idModel{}, IDFieldName: "UID"})
	repo := NewGenericRepositoryWithResource(db, res).(*GenericRepository)
	if repo.GetIDFieldName() != "UID" {
		t.Errorf("expected custom ID field name")
	}

	repo2 := NewGenericRepository(db, &idModel{}).(*GenericRepository)
	if repo2.GetIDFieldName() != "id" {
		t.Errorf("expected default id field name")
	}
}
