package main

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"unauthenticated", status.Error(codes.Unauthenticated, "invalid API key"), true},
		{"wrapped", fmt.Errorf("updating cache %w", status.Error(codes.Unauthenticated, "x")), true},
		{"other code", status.Error(codes.Unavailable, "down"), false},
		{"plain", errors.New("boom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAuthError(tt.err); got != tt.want {
				t.Errorf("isAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}
