package domain

import "context"

type ActionRepository interface {
	GetAllActions(ctx context.Context) ([]Action, error)
}

type ResourceRepository interface {
	GetAllResources(ctx context.Context) ([]Resource, error)
	Upsert(ctx context.Context, resource *Resource) error
}

type PolicyRepository interface {
	GetPoliciesPaging(ctx context.Context, offset, limit int) ([]Policy, error)
	Count(ctx context.Context) (int64, error)
}
