package domain

import "context"

type ActionRepository interface {
	GetAllActions(ctx context.Context) ([]Action, error)
	GetActionById(ctx context.Context, id int) (Action, error)
}

type ResourceRepository interface {
	GetAllResources(ctx context.Context) ([]Resource, error)
	Upsert(ctx context.Context, resource *Resource) error
	GetResourceById(ctx context.Context, id int) (Resource, error)
}

type PolicyRepository interface {
	GetPoliciesPaging(ctx context.Context, search string, offset, limit int) ([]Policy, error)
	Count(ctx context.Context, search string) (int64, error)
	Insert(ctx context.Context, policy *Policy) error
	GetPolicyById(ctx context.Context, id int) (Policy, error)
}

type AuthRepository interface {
	FindByUsername(ctx context.Context, username string) (Auth, error)
}
