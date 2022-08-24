package client

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/charm/kv"
)

type Entry struct {
	Key  []byte
	Data string
}

func TimeToKey(t time.Time) []byte {
	return []byte(t.Format(time.RFC3339))
}

func SaveEntry(ctx context.Context, db *kv.KV, when time.Time, text string) error {
	return db.Set(TimeToKey(when), text)
}

func DeleteEntry(ctx context.Context, db *kv.KV, when time.Time) error {
	return fmt.Errorf("unimplemented")
}

func FindNearestKey(ctx context.Context, db *kv.KV, when time.Time) []byte {
	return nil
}

func GetEntry(ctx context.Context, db *kv.KV, when time.Time) (*Entry, error) {
	return nil, fmt.Errorf("unimplemented")
}

func ListEntries(ctx context.Context, db *kv.KV, count int64) ([]*Entry, error) {
	return nil, fmt.Errorf("unimplemented")
}
