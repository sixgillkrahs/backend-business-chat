package domain

import "context"

type ActionRepository interface {
	GetAllActions(ctx context.Context) ([]Action, error)
}

type ResourceRepository interface {
	GetAllResources(ctx context.Context) ([]Resource, error)
	Upsert(ctx context.Context, resource *Resource) error
}
