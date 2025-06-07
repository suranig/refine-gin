package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
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
	ID         uint      `json:"id" gorm:"primaryKey"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CategoryID *uint     `json:"categoryId" gorm:"index"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Category   *Category `json:"-" gorm:"foreignKey:CategoryID" relation:"resource=categories;type=many-to-one;field=category"`
	Tags       []Tag     `json:"tags" gorm:"many2many:post_tags;" relation:"resource=tags;type=many-to-many;field=tags"`
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

// Category model
type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Posts     []Post    `json:"posts" gorm:"foreignKey:CategoryID" relation:"resource=posts;type=one-to-many;field=posts"`
}

// Tag model
type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Posts     []Post    `json:"posts" gorm:"many2many:post_tags;" relation:"resource=posts;type=many-to-many;field=posts"`
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

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Category{}, &Post{}, &Tag{})
	if err != nil {
		return nil, err
	}

	categories := []Category{
		{Name: "Technology"},
		{Name: "Health"},
		{Name: "Business"},
	}

	for i := range categories {
		db.Create(&categories[i])
	}

	tags := []Tag{
		{Name: "Go"},
		{Name: "Web"},
		{Name: "API"},
		{Name: "Database"},
	}

	for i := range tags {
		db.Create(&tags[i])
	}

	posts := []Post{
		{
			Title:      "Introduction to Go",
			Content:    "Go is a statically typed language developed by Google.",
			CategoryID: &categories[0].ID,
		},
		{
			Title:      "Health Benefits of Exercise",
			Content:    "Regular exercise has many health benefits.",
			CategoryID: &categories[1].ID,
		},
		{
			Title:      "Business Strategies",
			Content:    "Effective business strategies for growth.",
			CategoryID: &categories[2].ID,
		},
	}

	for i := range posts {
		db.Create(&posts[i])
	}

	db.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", posts[0].ID, tags[0].ID)
	db.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", posts[0].ID, tags[1].ID)
	db.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", posts[1].ID, tags[2].ID)
	db.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", posts[2].ID, tags[3].ID)

	return db, nil
}

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	api := r.Group("/api")

	categoryResource := resource.NewResource(resource.ResourceConfig{
		Name:  "categories",
		Model: Category{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
			resource.OperationCustom,
		},
	})

	postResource := resource.NewResource(resource.ResourceConfig{
		Name:  "posts",
		Model: Post{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
			resource.OperationCustom,
		},
	})

	tagResource := resource.NewResource(resource.ResourceConfig{
		Name:  "tags",
		Model: Tag{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
			resource.OperationCustom,
		},
	})

	categoryRepo := repository.NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &Category{},
	}))

	postRepo := repository.NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &Post{},
	}))

	tagRepo := repository.NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &Tag{},
	}))

	handler.RegisterResourceForRefineWithRelations(api, categoryResource, categoryRepo, "id", []string{"posts"})
	handler.RegisterResourceForRefineWithRelations(api, postResource, postRepo, "id", []string{"category", "tags"})
	handler.RegisterResourceForRefineWithRelations(api, tagResource, tagRepo, "id", []string{"posts"})

	publishAction := handler.CustomAction{
		Name:       "publish",
		Method:     "POST",
		RequiresID: true,
		Handler: func(c *gin.Context, res resource.Resource, repo repository.Repository) (interface{}, error) {
			id := c.Param("id")
			post, err := repo.Get(c, id)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": "Post has been published",
				"post":    post,
			}, nil
		},
	}

	handler.RegisterCustomActions(api, postResource, postRepo, []handler.CustomAction{publishAction})

	log.Println("Server running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
