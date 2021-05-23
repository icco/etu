package etu

import (
	"fmt"
	"os"
)

type EmailRequest struct {
	From          string `json:"From"`
	MessageStream string `json:"MessageStream"`
	FromName      string `json:"FromName"`
	FromFull      struct {
		Email       string `json:"Email"`
		Name        string `json:"Name"`
		MailboxHash string `json:"MailboxHash"`
	} `json:"FromFull"`
	To     string `json:"To"`
	ToFull []struct {
		Email       string `json:"Email"`
		Name        string `json:"Name"`
		MailboxHash string `json:"MailboxHash"`
	} `json:"ToFull"`
	Cc     string `json:"Cc"`
	CcFull []struct {
		Email       string `json:"Email"`
		Name        string `json:"Name"`
		MailboxHash string `json:"MailboxHash"`
	} `json:"CcFull"`
	Bcc     string `json:"Bcc"`
	BccFull []struct {
		Email       string `json:"Email"`
		Name        string `json:"Name"`
		Mailboxhash string `json:"MailboxHash"`
	} `json:"BccFull"`
	OriginalRecipient string `json:"OriginalRecipient"`
	ReplyTo           string `json:"ReplyTo"`
	Subject           string `json:"Subject"`
	MessageID         string `json:"MessageID"`
	Date              string `json:"Date"`
	MailboxHash       string `json:"MailboxHash"`
	TextBody          string `json:"TextBody"`
	HTMLBody          string `json:"HtmlBody"`
	StrippedTextReply string `json:"StrippedTextReply"`
	Tag               string `json:"Tag"`
	Headers           []struct {
		Name  string `json:"Name"`
		Value string `json:"Value"`
	} `json:"Headers"`
	Attachments []struct {
		Name          string `json:"Name"`
		Content       string `json:"Content"`
		ContentType   string `json:"ContentType"`
		ContentLength int    `json:"ContentLength"`
		ContentID     string `json:"ContentID"`
	} `json:"Attachments"`
}

// Validate checks if the request was valid.
func (r *EmailRequest) Validate() error {
	if req.FromFull.Email != os.Getenv("EXPECTED_FROM_EMAIL") {
		return fmt.Errorf("%q is not a valid from email")
	}

	if req.ToFull.Email != os.Getenv("EXPECTED_TO_EMAIL") {
		return fmt.Errorf("%q is not a valid to email")
	}

	return nil
}

// Save uploads to graphql.
func (r *EmailRequest) Save(ctx context.Context, client *graphql.Client) error {
	return EditPage(ctx, client, r.Subject, r.TextBody, nil)
}
