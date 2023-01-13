package client

import (
	"context"
	"database/sql"
	"sort"
	"time"
)

type Post struct {
	"gorm.io/gorm"

	ID        string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func TimeToKey(t time.Time) string {
	return t.Format(time.RFC3339)
}

func Sync(db *sql.DB) error {
	return nil
}

func set(db *sql.DB, key, value string) error {
	return nil
}

func get(db *sql.DB, key string) (string, error) {
	return "", nil
}

func keys(db *sql.DB) ([]string, error) {
	return nil, nil
}

func delete(db *sql.DB, key string) error {
	return nil
}

func TimeSinceLastPost(ctx context.Context, db *sql.DB) (time.Duration, error) {
	return 0, nil
}

func SaveEntry(ctx context.Context, db *sql.DB, when time.Time, text string) error {
	return set(db, TimeToKey(when), text)
}

func DeleteEntry(ctx context.Context, db *sql.DB, key string) error {
	return delete(db, key)
}

func FindNearestKey(ctx context.Context, db *sql.DB, when time.Time) (string, error) {
	return "", nil
}

func GetEntry(ctx context.Context, db *sql.DB, key string) (*Entry, error) {
	d, err := get(db, key)
	if err != nil {
		return nil, err
	}

	return &Entry{
		Data: d,
		Key:  key,
	}, nil
}

func ListEntries(ctx context.Context, db *sql.DB, count int) ([]*Entry, error) {
	keys, err := keys(db)
	if err != nil {
		return nil, err
	}

	sort.Slice(keys, func(i, j int) bool {
		return string(keys[j]) < string(keys[i])
	})

	var entries []*Entry
	for i := 0; i < count && i < len(keys); i++ {
		e, err := GetEntry(ctx, db, keys[i])
		if err != nil {
			return nil, err
		}

		entries = append(entries, e)
	}

	return entries, nil
}
