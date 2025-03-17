package repository

import (
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// GormRepositoryFactory implements the RepositoryFactory for GORM
type GormRepositoryFactory struct {
	DB *gorm.DB
}

// CreateRepository creates a new repository for a resource
func (f *GormRepositoryFactory) CreateRepository(res resource.Resource) Repository {
	return NewGormRepositoryWithResource(f.DB, res)
}

// NewGormRepositoryFactory creates a new GormRepositoryFactory
func NewGormRepositoryFactory(db *gorm.DB) RepositoryFactory {
	return &GormRepositoryFactory{
		DB: db,
	}
}
