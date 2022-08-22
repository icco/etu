package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/icco/etu/models"
	"github.com/tidwall/buntdb"
)

func GetKey(e *models.Entry) string {
	return fmt.Sprintf("user:%s:entries:%d", e.User.ID, e.DateTime.UnixMicro())
}

func SaveEntry(ctx context.Context, db *buntdb.DB, e *models.Entry) error {
	return db.Update(func(tx *buntdb.Tx) error {
		bts, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshal Entry: %w", err)
		}

		_, _, err := tx.Set(GetKey(e), string(bts), nil)
		return err
	})
}
