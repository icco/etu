package client

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{"empty key", "", true},
		{"valid key", "test-key-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{ApiKey: tt.apiKey}
			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectMIME(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		path string
		want string
	}{
		{
			"jpeg magic bytes",
			[]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			"photo.jpg",
			"image/jpeg",
		},
		{
			"png magic bytes",
			[]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			"image.png",
			"image/png",
		},
		{
			"extension fallback",
			[]byte{0x00, 0x00, 0x00, 0x00},
			"image.png",
			"image/png",
		},
		{
			"text content",
			[]byte("hello world"),
			"file.txt",
			"text/plain; charset=utf-8",
		},
		{
			"unknown content and extension",
			[]byte{0x00, 0x00, 0x00, 0x00},
			"file.nope",
			"application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectMIME(tt.data, tt.path)
			if got != tt.want {
				t.Errorf("detectMIME() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadImageUploads(t *testing.T) {
	t.Run("nil paths", func(t *testing.T) {
		uploads, err := LoadImageUploads(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uploads != nil {
			t.Errorf("expected nil, got %v", uploads)
		}
	})

	t.Run("empty paths", func(t *testing.T) {
		uploads, err := LoadImageUploads([]string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uploads != nil {
			t.Errorf("expected nil, got %v", uploads)
		}
	})

	t.Run("valid png file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.png")
		// PNG magic bytes
		data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}

		uploads, err := LoadImageUploads([]string{path})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uploads) != 1 {
			t.Fatalf("expected 1 upload, got %d", len(uploads))
		}
		if uploads[0].MimeType != "image/png" {
			t.Errorf("MimeType = %q, want %q", uploads[0].MimeType, "image/png")
		}
		if len(uploads[0].Data) != len(data) {
			t.Errorf("Data length = %d, want %d", len(uploads[0].Data), len(data))
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		dir := t.TempDir()
		paths := make([]string, 3)
		for i := range paths {
			p := filepath.Join(dir, filepath.Base(t.Name())+string(rune('a'+i))+".jpg")
			if err := os.WriteFile(p, []byte{0xFF, 0xD8, 0xFF}, 0644); err != nil {
				t.Fatal(err)
			}
			paths[i] = p
		}

		uploads, err := LoadImageUploads(paths)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uploads) != 3 {
			t.Fatalf("expected 3 uploads, got %d", len(uploads))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadImageUploads([]string{"/nonexistent/file.png"})
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestLoadAudioUploads(t *testing.T) {
	t.Run("nil paths", func(t *testing.T) {
		uploads, err := LoadAudioUploads(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uploads != nil {
			t.Errorf("expected nil, got %v", uploads)
		}
	})

	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.wav")
		if err := os.WriteFile(path, []byte("fake audio data"), 0644); err != nil {
			t.Fatal(err)
		}

		uploads, err := LoadAudioUploads([]string{path})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uploads) != 1 {
			t.Fatalf("expected 1 upload, got %d", len(uploads))
		}
		if uploads[0].MimeType == "" {
			t.Error("expected non-empty MIME type")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadAudioUploads([]string{"/nonexistent/file.mp3"})
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestCacheRoundTrip(t *testing.T) {
	setTestHome(t)

	cfg := &Config{ApiKey: "test"}
	dur := 5 * time.Minute

	if err := cfg.cacheToFile(dur); err != nil {
		t.Fatalf("cacheToFile: %v", err)
	}

	data, err := cfg.cacheFromFile()
	if err != nil {
		t.Fatalf("cacheFromFile: %v", err)
	}
	if data == nil {
		t.Fatal("expected cache data, got nil")
	}
	if data.Duration != dur {
		t.Errorf("Duration = %v, want %v", data.Duration, dur)
	}
	if time.Since(data.Saved) > time.Second {
		t.Errorf("Saved time too old: %v", data.Saved)
	}
}

func TestCacheOverwrite(t *testing.T) {
	setTestHome(t)

	cfg := &Config{ApiKey: "test"}

	if err := cfg.cacheToFile(1 * time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := cfg.cacheToFile(10 * time.Minute); err != nil {
		t.Fatal(err)
	}

	data, err := cfg.cacheFromFile()
	if err != nil {
		t.Fatal(err)
	}
	if data.Duration != 10*time.Minute {
		t.Errorf("Duration = %v, want %v", data.Duration, 10*time.Minute)
	}
}

func TestCacheFromFileMissing(t *testing.T) {
	setTestHome(t)

	cfg := &Config{ApiKey: "test"}
	data, err := cfg.cacheFromFile()
	if err == nil {
		t.Error("expected error for missing cache file")
	}
	if data != nil {
		t.Error("expected nil data for missing cache file")
	}
}
