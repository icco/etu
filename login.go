package main

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"go.icco.me/etu/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Save an API key to the config file.",
		Args:  cobra.NoArgs,
		RunE:  login,
	}

	logoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved API key from the config file.",
		Args:  cobra.NoArgs,
		RunE:  logout,
	}
)

func login(cmd *cobra.Command, _ []string) error {
	var key string
	err := huh.NewInput().
		Title("API key").
		Description(fmt.Sprintf("Create one at %s/settings/api", client.SignupURL)).
		EchoMode(huh.EchoModePassword).
		Value(&key).
		Validate(func(v string) error {
			if strings.TrimSpace(v) == "" {
				return errors.New("API key is required")
			}
			return nil
		}).
		Run()
	if err != nil {
		return err
	}

	check := *cfg
	check.APIKey = strings.TrimSpace(key)
	if _, err := check.ListPosts(cmd.Context(), 1); err != nil {
		if isAuthError(err) {
			return errors.New("API key rejected, check it and try again")
		}
		return fmt.Errorf("verifying API key: %w", err)
	}

	if err := client.SaveAPIKey(key); err != nil {
		return err
	}
	fmt.Println("Logged in.")
	return nil
}

func logout(_ *cobra.Command, _ []string) error {
	if err := client.SaveAPIKey(""); err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
}

func isAuthError(err error) bool {
	var se interface{ GRPCStatus() *status.Status }
	if !errors.As(err, &se) {
		return false
	}
	return se.GRPCStatus().Code() == codes.Unauthenticated
}
