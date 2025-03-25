package repository

import (
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// RepositoryFactory defines the interface for creating repositories
type RepositoryFactory interface {
	CreateRepository(res resource.Resource) Repository
}

// GenericRepositoryFactory implements the RepositoryFactory interface
type GenericRepositoryFactory struct {
	DB *gorm.DB
}

// CreateRepository creates a new generic repository for a resource
func (f *GenericRepositoryFactory) CreateRepository(res resource.Resource) Repository {
	return NewGenericRepositoryWithResource(f.DB, res)
}

// NewGenericRepositoryFactory creates a new GenericRepositoryFactory
func NewGenericRepositoryFactory(db *gorm.DB) RepositoryFactory {
	return &GenericRepositoryFactory{
		DB: db,
	}
}
