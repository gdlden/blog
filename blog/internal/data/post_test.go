package data

import (
	"context"
	"testing"

	"blog/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	// Auto migrate the schema
	err = db.AutoMigrate(&Post{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestPostRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := &postRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()
	post := &biz.Post{
		Title:   "Test Title",
		Content: "Test Content",
	}

	result, err := repo.Save(ctx, post)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Title", result.Title)
	assert.Equal(t, "Test Content", result.Content)
}

func TestPostRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := &postRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// First create a post
	post := &biz.Post{
		Title:   "Original Title",
		Content: "Original Content",
	}
	saved, _ := repo.Save(ctx, post)

	// Get the ID from database (it should be "1" for first record)
	var dbPost Post
	db.First(&dbPost)
	saved.Id = "1"

	// Update the post
	saved.Title = "Updated Title"
	saved.Content = "Updated Content"
	result, err := repo.Update(ctx, saved)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Title", result.Title)

	// Verify in database
	var updatedPost Post
	db.First(&updatedPost, 1)
	assert.Equal(t, "Updated Title", updatedPost.Title)
	assert.Equal(t, "Updated Content", updatedPost.Content)
}

func TestPostRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := &postRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a post first
	post := &biz.Post{
		Title:   "To Delete",
		Content: "Content",
	}
	repo.Save(ctx, post)

	// Get the ID
	var dbPost Post
	db.First(&dbPost)

	// Delete the post
	err := repo.Delete(ctx, int64(dbPost.ID))

	assert.NoError(t, err)

	// Verify deletion
	var count int64
	db.Model(&Post{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestPostRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := &postRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a post
	post := &biz.Post{
		Title:   "Find Me",
		Content: "Content",
	}
	repo.Save(ctx, post)

	// Get the ID
	var dbPost Post
	db.First(&dbPost)

	// Find by ID
	result, err := repo.FindByID(ctx, int64(dbPost.ID))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Find Me", result.Title)
	assert.Equal(t, "Content", result.Content)
}

func TestPostRepo_FindByPage(t *testing.T) {
	db := setupTestDB(t)
	repo := &postRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create multiple posts
	for i := 0; i < 5; i++ {
		post := &biz.Post{
			Title:   "Post",
			Content: "Content",
		}
		repo.Save(ctx, post)
	}

	// Test pagination
	posts, total, err := repo.FindByPage(ctx, &biz.PostPageRequest{
		Current: 1,
		Size:    3,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, posts, 3)
}
