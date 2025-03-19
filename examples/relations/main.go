package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User model
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Posts     []Post    `json:"posts" gorm:"foreignKey:AuthorID" relation:"resource=posts;type=one-to-many;field=author_id;reference=id;include=false"`
	Profile   *Profile  `json:"profile" gorm:"foreignKey:UserID" relation:"resource=profiles;type=one-to-one;field=user_id;reference=id;include=true"`
}

// Post model
type Post struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id" gorm:"index"`
	Author    User      `json:"author" gorm:"foreignKey:AuthorID" relation:"resource=users;type=many-to-one;field=author_id;reference=id;include=true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `json:"comments" gorm:"foreignKey:PostID" relation:"resource=comments;type=one-to-many;field=post_id;reference=id;include=false"`
}

// Comment model
type Comment struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id" gorm:"index"`
	Author    User      `json:"author" gorm:"foreignKey:AuthorID" relation:"resource=users;type=many-to-one;field=author_id;reference=id;include=true"`
	PostID    string    `json:"post_id" gorm:"index"`
	Post      Post      `json:"post" gorm:"foreignKey:PostID" relation:"resource=posts;type=many-to-one;field=post_id;reference=id;include=false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Profile model
type Profile struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Bio       string    `json:"bio"`
	UserID    string    `json:"user_id" gorm:"uniqueIndex"`
	User      User      `json:"user" gorm:"foreignKey:UserID" relation:"resource=users;type=one-to-one;field=user_id;reference=id;include=false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository implements repository for User
type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var users []User
	var total int64

	q := r.db.Model(&User{})

	// Apply includes from options
	includes := resource.IncludeRelations(ctx.(*gin.Context), options.Resource)
	for _, include := range includes {
		q = q.Preload(include)
	}

	total, err := options.ApplyWithPagination(q, &users)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var user User

	// Get includes from context
	includes := resource.IncludeRelations(ctx.(*gin.Context), ctx.Value("resource").(resource.Resource))

	// Apply includes
	q := r.db
	for _, include := range includes {
		q = q.Preload(include)
	}

	if err := q.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	user := data.(*User)

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	user := data.(*User)
	user.ID = id.(string)
	user.UpdatedAt = time.Now()

	if err := r.db.Save(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id interface{}) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}

// Count returns the total number of resources matching the query options
func (r *UserRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	var count int64
	db := r.db.Model(&User{})

	// Apply filters from options
	db = options.Apply(db)

	// Get count
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CreateMany creates multiple users
func (r *UserRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	users, ok := data.([]User)
	if !ok {
		return nil, fmt.Errorf("invalid data type, expected []User")
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&users).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateMany updates multiple users
func (r *UserRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	user, ok := data.(User)
	if !ok {
		return 0, fmt.Errorf("invalid data type, expected User")
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&User{}).Where("id IN ?", ids).Updates(user)
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}

// DeleteMany deletes multiple users
func (r *UserRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Where("id IN ?", ids).Delete(&User{})
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}

// Similar repositories for Post, Comment, and Profile...

func main() {
	// Setup database
	db, err := gorm.Open(sqlite.Open("relations.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate models
	err = db.AutoMigrate(&User{}, &Post{}, &Comment{}, &Profile{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create repositories
	userRepo := &UserRepository{db: db}
	// Create other repositories...

	// Create resources
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		DefaultSort: &resource.Sort{
			Field: "created_at",
			Order: "desc",
		},
	})

	// Create other resources...

	// Setup Gin
	r := gin.Default()
	api := r.Group("/api")

	// Register resources
	handler.RegisterResource(api, userResource, userRepo)
	// Register other resources...

	// Start server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
