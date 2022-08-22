package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/icco/etu/graph/generated"
	"github.com/icco/etu/models"
)

// CreateEntry is the resolver for the createEntry field.
func (r *mutationResolver) CreateEntry(ctx context.Context, input models.NewEntry) (*models.Entry, error) {
	panic(fmt.Errorf("not implemented: CreateEntry - createEntry"))
}

// Entries is the resolver for the entries field.
func (r *queryResolver) Entries(ctx context.Context) ([]*models.Entry, error) {
	panic(fmt.Errorf("not implemented: Entries - entries"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
