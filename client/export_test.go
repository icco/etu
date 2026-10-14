package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"go.icco.me/etu-backend/proto"
)

func feed(items []*proto.ExportDataResponse, final error) func() (*proto.ExportDataResponse, error) {
	i := 0
	return func() (*proto.ExportDataResponse, error) {
		if i < len(items) {
			i++
			return items[i-1], nil
		}
		return nil, final
	}
}

func TestWriteExport(t *testing.T) {
	items := []*proto.ExportDataResponse{
		{Item: &proto.ExportDataResponse_User{User: &proto.User{Email: "ada@example.com"}}},
		{Item: &proto.ExportDataResponse_Tag{Tag: &proto.Tag{Name: "math"}}},
		{Item: &proto.ExportDataResponse_Note{Note: &proto.Note{Id: "n1", Content: "engine notes"}}},
	}
	var buf bytes.Buffer
	n, err := writeExport(&buf, feed(items, io.EOF))
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("n = %d, want 3", n)
	}
	var got []map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(got) != 3 || got[0]["user"] == nil || got[1]["tag"] == nil || got[2]["note"] == nil {
		t.Errorf("unexpected items: %s", buf.String())
	}
}

func TestWriteExportEmpty(t *testing.T) {
	var buf bytes.Buffer
	if _, err := writeExport(&buf, feed(nil, io.EOF)); err != nil {
		t.Fatal(err)
	}
	var got []any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil || len(got) != 0 {
		t.Errorf("got %v, %v from %q", got, err, buf.String())
	}
}

func TestWriteExportStreamError(t *testing.T) {
	boom := errors.New("boom")
	if _, err := writeExport(io.Discard, feed(nil, boom)); !errors.Is(err, boom) {
		t.Errorf("err = %v, want boom", err)
	}
}
