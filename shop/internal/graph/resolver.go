package graph

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ShoppingDem/backend/shop/pkg/models"
)

type Resolver struct {
	DB *sql.DB
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	// Implement user creation logic here
	// This should include Okta registration and database insertion
	return nil, errors.New("not implemented")
}

func (r *mutationResolver) Login(ctx context.Context, input LoginInput) (string, error) {
	// Implement login logic here
	// This should include Okta validation
	return "", errors.New("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) User(ctx context.Context, id string) (*models.User, error) {
	// Implement user retrieval logic here
	return nil, errors.New("not implemented")
}
