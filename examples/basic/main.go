package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Sample data
var users []User
var posts []Post

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" refine:"filterable;searchable"`
	Email     string    `json:"email" refine:"filterable"`
	CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var users []User
	var total int64

	q := r.db.Model(&User{})
	total, err := options.ApplyWithPagination(q, &users)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	user := data.(*User)
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	user := data.(*User)
	user.ID = id.(string)

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
	query := r.db.Model(&User{})

	// Apply filters
	for field, value := range options.Filters {
		query = query.Where(field+" = ?", value)
	}

	// Apply search
	if options.Search != "" {
		query = query.Where("name LIKE ?", "%"+options.Search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *UserRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	usersArray, ok := data.([]User)
	if !ok {
		return nil, errors.New("data must be []User")
	}

	for i := range usersArray {
		usersArray[i].ID = uuid.New().String()
		users = append(users, usersArray[i])
	}

	return usersArray, nil
}

func (r *UserRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	user, ok := data.(User)
	if !ok {
		return 0, errors.New("data must be User")
	}

	count := int64(0)
	for _, id := range ids {
		for i, u := range users {
			if u.ID == id.(string) {
				// Update user fields while preserving ID
				user.ID = u.ID
				users[i] = user
				count++
				break
			}
		}
	}

	return count, nil
}

func (r *UserRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	count := int64(0)
	var newUsers []User

	for _, u := range users {
		found := false
		for _, id := range ids {
			if u.ID == id.(string) {
				found = true
				count++
				break
			}
		}
		if !found {
			newUsers = append(newUsers, u)
		}
	}

	users = newUsers
	return count, nil
}

type Post struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" refine:"filterable;searchable"`
	Content   string    `json:"content"`
	UserID    string    `json:"user_id" refine:"filterable"`
	CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

type PostRepository struct {
	db *gorm.DB
}

func (r *PostRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var posts []Post
	var total int64

	q := r.db.Model(&Post{})
	total, err := options.ApplyWithPagination(q, &posts)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *PostRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var post Post
	if err := r.db.First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (r *PostRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	post := data.(*Post)
	if err := r.db.Create(post).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (r *PostRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	post := data.(*Post)
	post.ID = id.(string)

	if err := r.db.Save(post).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (r *PostRepository) Delete(ctx context.Context, id interface{}) error {
	return r.db.Delete(&Post{}, "id = ?", id).Error
}

// Count returns the total number of resources matching the query options
func (r *PostRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	var count int64
	query := r.db.Model(&Post{})

	// Apply filters
	for field, value := range options.Filters {
		query = query.Where(field+" = ?", value)
	}

	// Apply search
	if options.Search != "" {
		query = query.Where("title LIKE ?", "%"+options.Search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	postsArray, ok := data.([]Post)
	if !ok {
		return nil, errors.New("data must be []Post")
	}

	for i := range postsArray {
		postsArray[i].ID = uuid.New().String()
		posts = append(posts, postsArray[i])
	}

	return postsArray, nil
}

func (r *PostRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	post, ok := data.(Post)
	if !ok {
		return 0, errors.New("data must be Post")
	}

	count := int64(0)
	for _, id := range ids {
		for i, p := range posts {
			if p.ID == id.(string) {
				// Update post fields while preserving ID
				post.ID = p.ID
				posts[i] = post
				count++
				break
			}
		}
	}

	return count, nil
}

func (r *PostRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	count := int64(0)
	var newPosts []Post

	for _, p := range posts {
		found := false
		for _, id := range ids {
			if p.ID == id.(string) {
				found = true
				count++
				break
			}
		}
		if !found {
			newPosts = append(newPosts, p)
		}
	}

	posts = newPosts
	return count, nil
}

func main() {
	// Create a new Gin router
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&User{}, &Post{})

	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		users := []User{
			{ID: "1", Name: "John Doe", Email: "john@example.com", CreatedAt: time.Now()},
			{ID: "2", Name: "Jane Smith", Email: "jane@example.com", CreatedAt: time.Now()},
		}

		for _, user := range users {
			db.Create(&user)
		}

		posts := []Post{
			{ID: "1", Title: "First Post", Content: "This is the first post", UserID: "1", CreatedAt: time.Now()},
			{ID: "2", Title: "Second Post", Content: "This is the second post", UserID: "1", CreatedAt: time.Now()},
			{ID: "3", Title: "Jane's Post", Content: "This is Jane's post", UserID: "2", CreatedAt: time.Now()},
		}

		for _, post := range posts {
			db.Create(&post)
		}
	}

	userRepo := &UserRepository{db: db}
	postRepo := &PostRepository{db: db}

	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		DefaultSort: &resource.Sort{
			Field: "created_at",
			Order: string(query.SortOrderDesc),
		},
	})

	postResource := resource.NewResource(resource.ResourceConfig{
		Name:  "posts",
		Model: Post{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Register resources
	api := r.Group("/api")
	handler.RegisterResource(api, userResource, userRepo)
	handler.RegisterResource(api, postResource, postRepo)

	// Configure Swagger info
	swaggerInfo := swagger.SwaggerInfo{
		Title:       "Basic Example API",
		Description: "API documentation for the basic example",
		Version:     "1.0.0",
		BasePath:    "/api",
	}

	// Register Swagger routes
	swagger.RegisterSwagger(r.Group(""), []resource.Resource{userResource, postResource}, swaggerInfo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start the server
	log.Printf("Server starting on http://localhost:9003")
	if err := r.Run(":9003"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
