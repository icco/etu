package client

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/kirsle/configdir"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	dbFilename = "etu.db"
	appName    = "etu"
)

type Post struct {
	gorm.Model

	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Content   string
	CreatedAt time.Time `sql:"index"`
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// BeforeCreate will set a UUID as the primary key.
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	p.ID = uuid
	p.CreatedAt = time.Now()
	p.DeletedAt = nil
	p.UpdatedAt = time.Now()

	return nil
}

func (p *Post) BeforeSave(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()

	return nil
}

func openDB() (*gorm.DB, error) {
	configPath := configdir.LocalConfig(appName)
	if err := configdir.MakePath(configPath); err != nil {
		return nil, err
	}

	dbFile := filepath.Join(configPath, dbFilename)

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(&Post{}); err != nil {
		return nil, err
	}

	return db, nil
}

func Sync(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func TimeSinceLastPost(ctx context.Context) (time.Duration, error) {
	posts, err := ListPosts(ctx, 1)
	if err != nil {
		return 0, err
	}

	if len(posts) != 1 {
		return 0, fmt.Errorf("incorrect number of posts found")
	}

	dur := time.Now().Sub(posts[0].CreatedAt)
	return dur, nil
}

func SaveEntry(ctx context.Context, text string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	p := &Post{
		Content: text,
	}
	return db.WithContext(ctx).Create(p).Error
}

func DeletePost(ctx context.Context, key string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Delete(&Post{}, key).Error
}

func GetPost(ctx context.Context, key string) (*Post, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}

	var post *Post
	if err := db.WithContext(ctx).Find(&post, key).Error; err != nil {
		return nil, err
	}

	return post, nil
}

func ListPosts(ctx context.Context, count int) ([]*Post, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	var posts []*Post
	if err := db.WithContext(ctx).Order("created_at desc").Limit(count).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}
