package client

import (
	"bytes"
	"context"
	"io"
	"sort"
	"time"

	"github.com/charmbracelet/charm/crypt"
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
	cr, err := crypt.NewCrypt()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	eb, err := cr.NewEncryptedWriter(buf)
	if err != nil {
		return err
	}

	if _, err := io.WriteString(eb, text); err != nil {
		return err
	}
	eb.Close()

	return db.Set(TimeToKey(when), buf.Bytes())
}

func DeleteEntry(ctx context.Context, db *kv.KV, key []byte) error {
	return db.Delete(key)
}

func FindNearestKey(ctx context.Context, db *kv.KV, when time.Time) []byte {
	return nil
}

func GetEntry(ctx context.Context, db *kv.KV, key []byte) (*Entry, error) {
	d, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	cr, err := crypt.NewCrypt()
	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(d)
	deb, err := cr.NewDecryptedReader(br)
	if err != nil {
		return nil, err
	}

	decoded, err := io.ReadAll(deb)
	if err != nil {
		return nil, err
	}

	return &Entry{
		Data: string(decoded),
		Key:  key,
	}, nil
}

func ListEntries(ctx context.Context, db *kv.KV, count int) ([]*Entry, error) {
	keys, err := db.Keys()
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
