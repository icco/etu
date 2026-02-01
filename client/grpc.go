package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/icco/etu-backend/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// noteToPost converts a proto Note to a client Post for TUI/CLI use.
func noteToPost(n *proto.Note) *Post {
	if n == nil {
		return nil
	}
	var createdAt, updatedAt time.Time
	if t := n.GetCreatedAt(); t != nil {
		createdAt = t.AsTime()
	}
	if t := n.GetUpdatedAt(); t != nil {
		updatedAt = t.AsTime()
	}
	return &Post{
		ID:         n.GetId(),
		PageID:     n.GetId(),
		Tags:       n.GetTags(),
		Text:       n.GetContent(),
		CreatedAt:  createdAt,
		ModifiedAt: updatedAt,
	}
}

// apiKeyCreds attaches the etu API key to every gRPC request.
type apiKeyCreds struct {
	apiKey string
}

func (a apiKeyCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	key := a.apiKey
	if !strings.HasPrefix(key, "etu_") {
		key = "etu_" + key
	}
	return map[string]string{
		"authorization": key,
	}, nil
}

func (a apiKeyCreds) RequireTransportSecurity() bool {
	return true
}

// grpcClients holds the gRPC connection and service clients (lazy-init).
type grpcClients struct {
	conn          *grpc.ClientConn
	notesClient   proto.NotesServiceClient
	apiKeysClient proto.ApiKeysServiceClient
	userID        string
	connOnce      sync.Once
	userIDOnce    sync.Once
	connErr       error
	userIDErr     error
}

func (c *Config) getGRPCClients() (*grpcClients, error) {
	if c.ApiKey == "" {
		return nil, fmt.Errorf("API key not set")
	}
	// Lazy-init is done in ensureGRPCConn; we need a shared struct. Store on Config.
	if c.grpc == nil {
		c.grpc = &grpcClients{}
	}
	c.grpc.connOnce.Do(func() {
		creds := credentials.NewTLS(&tls.Config{})
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(apiKeyCreds{apiKey: c.ApiKey}),
		}
		c.grpc.conn, c.grpc.connErr = grpc.NewClient(c.GRPCTarget, opts...)
		if c.grpc.connErr != nil {
			return
		}
		c.grpc.notesClient = proto.NewNotesServiceClient(c.grpc.conn)
		c.grpc.apiKeysClient = proto.NewApiKeysServiceClient(c.grpc.conn)
	})
	if c.grpc.connErr != nil {
		return nil, c.grpc.connErr
	}
	return c.grpc, nil
}

// ensureUserID calls VerifyApiKey and caches user_id for use in note requests.
func (c *Config) ensureUserID(ctx context.Context) (string, error) {
	g, err := c.getGRPCClients()
	if err != nil {
		return "", err
	}
	g.userIDOnce.Do(func() {
		resp, err := g.apiKeysClient.VerifyApiKey(ctx, &proto.VerifyApiKeyRequest{
			RawKey: c.ApiKey,
		})
		if err != nil {
			g.userIDErr = err
			return
		}
		if !resp.GetValid() {
			g.userIDErr = fmt.Errorf("API key invalid")
			return
		}
		g.userID = resp.GetUserId()
	})
	if g.userIDErr != nil {
		return "", g.userIDErr
	}
	return g.userID, nil
}
