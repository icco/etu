package main

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	subscriptionURL = "https://etu.timeclimbers.com/settings/subscription"
	apiSettingsURL  = "https://etu.timeclimbers.com/settings/api"
)

// friendlyError turns backend auth and paywall errors into actionable messages.
func friendlyError(err error) error {
	var se interface{ GRPCStatus() *status.Status }
	if !errors.As(err, &se) {
		return err
	}
	st := se.GRPCStatus()
	switch st.Code() {
	case codes.PermissionDenied:
		if st.Message() == "subscription required" {
			return fmt.Errorf("an active subscription is required, manage it at %s", subscriptionURL)
		}
	case codes.Unauthenticated:
		return errors.New("API key rejected (keys are revoked when your password changes), create a new one at " + apiSettingsURL)
	}
	return err
}
