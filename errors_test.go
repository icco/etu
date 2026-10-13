package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFriendlyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"paywall", status.Error(codes.PermissionDenied, "subscription required"), subscriptionURL},
		{"wrapped paywall", fmt.Errorf("list tags: %w", status.Error(codes.PermissionDenied, "subscription required")), subscriptionURL},
		{"other permission denied", status.Error(codes.PermissionDenied, "nope"), "nope"},
		{"unauthenticated", status.Error(codes.Unauthenticated, "invalid API key"), apiKeysURL},
		{"plain error", errors.New("boom"), "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := friendlyError(tt.err).Error(); !strings.Contains(got, tt.want) {
				t.Errorf("friendlyError() = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}
