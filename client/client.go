package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

func DeleteEntry(ctx context.Context, db *kv.KV, when time.Time) error {
	return fmt.Errorf("unimplemented")
}

func FindNearestKey(ctx context.Context, db *kv.KV, when time.Time) []byte {
	return nil
}

func GetEntry(ctx context.Context, db *kv.KV, when time.Time) (*Entry, error) {
	key := TimeToKey(when)
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

func ListEntries(ctx context.Context, db *kv.KV, count int64) ([]*Entry, error) {
	return nil, fmt.Errorf("unimplemented")
}
