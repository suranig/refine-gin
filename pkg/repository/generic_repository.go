package repository

import (
	"context"

	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// GenericRepository to rozszerzona wersja GormRepository z dodatkowymi funkcjonalnościami
type GenericRepository struct {
	GormRepository
}

// NewGenericRepository tworzy nowe generyczne repozytorium na podstawie zasobu
func NewGenericRepository(db *gorm.DB, res resource.Resource) *GenericRepository {
	gormRepo := &GormRepository{
		DB:          db,
		Model:       resource.CreateInstanceOfType(res.GetModel()),
		Resource:    res,
		IDFieldName: res.GetIDFieldName(),
	}

	return &GenericRepository{
		GormRepository: *gormRepo,
	}
}

// ListWithRelations pobiera listę zasobów z uwzględnieniem relacji
func (r *GenericRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	slicePtr := resource.CreateSliceOfType(r.Model)

	// Przygotowanie zapytania
	query := r.DB.Model(r.Model)

	// Ładowanie relacji
	for _, relation := range relations {
		query = query.Preload(relation)
	}

	// Zastosowanie opcji zapytania
	total, err := options.ApplyWithPagination(query, slicePtr)
	if err != nil {
		return nil, 0, err
	}

	return slicePtr, total, nil
}

// GetWithRelations pobiera pojedynczy zasób z relacjami
func (r *GenericRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	item := resource.CreateInstanceOfType(r.Model)

	// Przygotowanie zapytania
	query := r.DB.Model(r.Model)

	// Ładowanie relacji
	for _, relation := range relations {
		query = query.Preload(relation)
	}

	// Wykonanie zapytania
	if err := query.First(item, r.IDFieldName+" = ?", id).Error; err != nil {
		return nil, err
	}

	return item, nil
}

// WithTransaction wykonuje operacje w transakcji
func (r *GenericRepository) WithTransaction(callback func(repo Repository) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Utwórz kopię repozytorium z transakcyjnym DB
		txRepo := &GenericRepository{
			GormRepository: GormRepository{
				DB:          tx,
				Model:       r.Model,
				Resource:    r.Resource,
				IDFieldName: r.IDFieldName,
			},
		}
		return callback(txRepo)
	})
}

// GenericRepositoryFactory to fabryka do tworzenia GenericRepository
type GenericRepositoryFactory struct {
	DB *gorm.DB
}

// CreateRepository tworzy nowe generyczne repozytorium dla zasobu
func (f *GenericRepositoryFactory) CreateRepository(res resource.Resource) Repository {
	return NewGenericRepository(f.DB, res)
}

// NewGenericRepositoryFactory tworzy nową fabrykę generycznych repozytoriów
func NewGenericRepositoryFactory(db *gorm.DB) RepositoryFactory {
	return &GenericRepositoryFactory{
		DB: db,
	}
}

// BulkCreate tworzy wiele rekordów jednocześnie
func (r *GenericRepository) BulkCreate(ctx context.Context, items interface{}) error {
	return r.DB.Create(items).Error
}

// BulkUpdate aktualizuje wiele rekordów jednocześnie na podstawie warunku
func (r *GenericRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	return r.DB.Model(r.Model).Where(condition).Updates(updates).Error
}

// Query zwraca query builder dla niestandardowych zapytań
func (r *GenericRepository) Query(ctx context.Context) *gorm.DB {
	return r.DB.Model(r.Model)
}

// FindOneBy znajduje pierwszy rekord spełniający warunek
func (r *GenericRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	item := resource.CreateInstanceOfType(r.Model)

	if err := r.DB.Where(condition).First(item).Error; err != nil {
		return nil, err
	}

	return item, nil
}

// FindAllBy znajduje wszystkie rekordy spełniające warunek
func (r *GenericRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	slicePtr := resource.CreateSliceOfType(r.Model)

	if err := r.DB.Where(condition).Find(slicePtr).Error; err != nil {
		return nil, err
	}

	return slicePtr, nil
}
