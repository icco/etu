package client

import (
	"context"
	"testing"
	"time"

	"github.com/icco/etu-backend/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestApiKeyCredsGetRequestMetadata(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"with prefix", "etu_abc123", "etu_abc123"},
		{"without prefix", "abc123", "etu_abc123"},
		{"empty key", "", "etu_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := apiKeyCreds{apiKey: tt.key}
			md, err := creds.GetRequestMetadata(context.Background())
			if err != nil {
				t.Fatalf("GetRequestMetadata: %v", err)
			}
			if got := md["authorization"]; got != tt.want {
				t.Errorf("authorization = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApiKeyCredsRequireTransportSecurity(t *testing.T) {
	creds := apiKeyCreds{apiKey: "test"}
	if !creds.RequireTransportSecurity() {
		t.Error("RequireTransportSecurity() = false, want true")
	}
}

func TestNoteToPost(t *testing.T) {
	t.Run("nil note", func(t *testing.T) {
		if got := noteToPost(nil); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("valid note", func(t *testing.T) {
		now := time.Now()
		note := &proto.Note{
			Id:        "test-id",
			Content:   "hello world",
			Tags:      []string{"tag1", "tag2"},
			CreatedAt: timestamppb.New(now),
		}

		got := noteToPost(note)
		if got == nil {
			t.Fatal("expected non-nil post")
		}
		if got.PageID != "test-id" {
			t.Errorf("PageID = %q, want %q", got.PageID, "test-id")
		}
		if got.Text != "hello world" {
			t.Errorf("Text = %q, want %q", got.Text, "hello world")
		}
		if len(got.Tags) != 2 || got.Tags[0] != "tag1" || got.Tags[1] != "tag2" {
			t.Errorf("Tags = %v, want [tag1 tag2]", got.Tags)
		}
		if got.CreatedAt.Sub(now).Abs() > time.Millisecond {
			t.Errorf("CreatedAt = %v, want ~%v", got.CreatedAt, now)
		}
	})

	t.Run("note without timestamp", func(t *testing.T) {
		note := &proto.Note{Id: "test-id", Content: "hello"}
		got := noteToPost(note)
		if !got.CreatedAt.IsZero() {
			t.Errorf("CreatedAt = %v, want zero", got.CreatedAt)
		}
	})

	t.Run("note with empty fields", func(t *testing.T) {
		note := &proto.Note{}
		got := noteToPost(note)
		if got == nil {
			t.Fatal("expected non-nil post for empty note")
		}
		if got.PageID != "" {
			t.Errorf("PageID = %q, want empty", got.PageID)
		}
	})
}

func TestNotesToPosts(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		got := notesToPosts(nil)
		if len(got) != 0 {
			t.Errorf("expected 0 posts, got %d", len(got))
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := notesToPosts([]*proto.Note{})
		if len(got) != 0 {
			t.Errorf("expected 0 posts, got %d", len(got))
		}
	})

	t.Run("multiple notes", func(t *testing.T) {
		notes := []*proto.Note{
			{Id: "1", Content: "first"},
			{Id: "2", Content: "second"},
			{Id: "3", Content: "third"},
		}
		got := notesToPosts(notes)
		if len(got) != 3 {
			t.Fatalf("expected 3 posts, got %d", len(got))
		}
		if got[0].PageID != "1" || got[1].PageID != "2" || got[2].PageID != "3" {
			t.Errorf("unexpected post IDs: %v, %v, %v", got[0].PageID, got[1].PageID, got[2].PageID)
		}
	})

	t.Run("filters nil notes", func(t *testing.T) {
		notes := []*proto.Note{
			{Id: "1", Content: "first"},
			nil,
			{Id: "3", Content: "third"},
		}
		got := notesToPosts(notes)
		if len(got) != 2 {
			t.Fatalf("expected 2 posts (nil filtered), got %d", len(got))
		}
	})
}

func TestGetGRPCClientsNoKey(t *testing.T) {
	c := &Config{}
	_, err := c.getGRPCClients()
	if err == nil {
		t.Error("expected error for empty API key")
	}
}
